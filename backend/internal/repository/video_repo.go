package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbvideocalltemplate "github.com/Wei-Shaw/sub2api/ent/videocalltemplate"
	dbvideogenerationtask "github.com/Wei-Shaw/sub2api/ent/videogenerationtask"
	dbvideomodel "github.com/Wei-Shaw/sub2api/ent/videomodel"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type videoRepository struct {
	client *dbent.Client
	db     *sql.DB
}

func NewVideoRepository(client *dbent.Client, db *sql.DB) service.VideoRepository {
	return &videoRepository{client: client, db: db}
}

func (r *videoRepository) GetModelByPublicModel(ctx context.Context, model string) (*service.VideoModel, error) {
	entModel, err := r.client.VideoModel.Query().
		Where(dbvideomodel.PublicModelEQ(strings.TrimSpace(model))).
		WithTemplate().
		Only(ctx)
	if dbent.IsNotFound(err) {
		return nil, service.ErrVideoModelNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get video model: %w", err)
	}
	return videoModelToService(entModel), nil
}

func (r *videoRepository) ListActiveModels(ctx context.Context) ([]service.VideoModel, error) {
	return r.ListModels(ctx, false)
}

func (r *videoRepository) ListModels(ctx context.Context, includeDisabled bool) ([]service.VideoModel, error) {
	q := r.client.VideoModel.Query().WithTemplate()
	if !includeDisabled {
		q = q.Where(dbvideomodel.StatusIn("active", "deprecated"))
	}
	entModels, err := q.Order(dbent.Asc(dbvideomodel.FieldSortOrder), dbent.Asc(dbvideomodel.FieldID)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list video models: %w", err)
	}
	out := make([]service.VideoModel, 0, len(entModels))
	for _, entModel := range entModels {
		out = append(out, *videoModelToService(entModel))
	}
	return out, nil
}

func (r *videoRepository) GetModelByID(ctx context.Context, id int64) (*service.VideoModel, error) {
	entModel, err := r.client.VideoModel.Query().
		Where(dbvideomodel.IDEQ(id)).
		WithTemplate().
		Only(ctx)
	if dbent.IsNotFound(err) {
		return nil, service.ErrVideoModelNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get video model by id: %w", err)
	}
	return videoModelToService(entModel), nil
}

func (r *videoRepository) GetTemplateByID(ctx context.Context, id int64) (*service.VideoCallTemplate, error) {
	entTpl, err := r.client.VideoCallTemplate.Get(ctx, id)
	if dbent.IsNotFound(err) {
		return nil, service.ErrVideoTemplateNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get video template: %w", err)
	}
	return videoTemplateToService(entTpl), nil
}

func (r *videoRepository) ListTemplates(ctx context.Context) ([]service.VideoCallTemplate, error) {
	entTemplates, err := r.client.VideoCallTemplate.Query().
		Order(dbent.Asc(dbvideocalltemplate.FieldName), dbent.Asc(dbvideocalltemplate.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list video templates: %w", err)
	}
	out := make([]service.VideoCallTemplate, 0, len(entTemplates))
	for _, entTpl := range entTemplates {
		out = append(out, *videoTemplateToService(entTpl))
	}
	return out, nil
}

func (r *videoRepository) CreateTemplate(ctx context.Context, template *service.VideoCallTemplate) error {
	builder := r.client.VideoCallTemplate.Create().
		SetName(template.Name).
		SetCreateMethod(template.CreateMethod).
		SetCreatePath(template.CreatePath).
		SetQueryMethod(template.QueryMethod).
		SetQueryPath(template.QueryPath).
		SetStatusMapping(template.StatusMapping).
		SetResultMapping(template.ResultMapping).
		SetErrorMapping(template.ErrorMapping).
		SetPollConfig(normalizeJSONMap(template.PollConfig)).
		SetTimeoutConfig(normalizeJSONMap(template.TimeoutConfig)).
		SetStatus(template.Status)
	if template.ContentMethod != nil {
		builder.SetContentMethod(*template.ContentMethod)
	}
	if template.ContentPath != nil {
		builder.SetContentPath(*template.ContentPath)
	}
	if template.CancelMethod != nil {
		builder.SetCancelMethod(*template.CancelMethod)
	}
	if template.CancelPath != nil {
		builder.SetCancelPath(*template.CancelPath)
	}
	entTpl, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create video template: %w", err)
	}
	*template = *videoTemplateToService(entTpl)
	return nil
}

func (r *videoRepository) UpdateTemplate(ctx context.Context, template *service.VideoCallTemplate) error {
	builder := r.client.VideoCallTemplate.UpdateOneID(template.ID).
		SetName(template.Name).
		SetCreateMethod(template.CreateMethod).
		SetCreatePath(template.CreatePath).
		SetQueryMethod(template.QueryMethod).
		SetQueryPath(template.QueryPath).
		SetStatusMapping(template.StatusMapping).
		SetResultMapping(template.ResultMapping).
		SetErrorMapping(template.ErrorMapping).
		SetPollConfig(normalizeJSONMap(template.PollConfig)).
		SetTimeoutConfig(normalizeJSONMap(template.TimeoutConfig)).
		SetStatus(template.Status)
	if template.ContentMethod != nil {
		builder.SetContentMethod(*template.ContentMethod)
	} else {
		builder.ClearContentMethod()
	}
	if template.ContentPath != nil {
		builder.SetContentPath(*template.ContentPath)
	} else {
		builder.ClearContentPath()
	}
	if template.CancelMethod != nil {
		builder.SetCancelMethod(*template.CancelMethod)
	} else {
		builder.ClearCancelMethod()
	}
	if template.CancelPath != nil {
		builder.SetCancelPath(*template.CancelPath)
	} else {
		builder.ClearCancelPath()
	}
	if _, err := builder.Save(ctx); err != nil {
		return fmt.Errorf("update video template: %w", err)
	}
	return nil
}

func (r *videoRepository) DeleteTemplate(ctx context.Context, id int64) error {
	if err := r.client.VideoCallTemplate.DeleteOneID(id).Exec(ctx); err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrVideoTemplateNotFound
		}
		return fmt.Errorf("delete video template: %w", err)
	}
	return nil
}

func (r *videoRepository) CreateModel(ctx context.Context, model *service.VideoModel) error {
	builder := r.client.VideoModel.Create().
		SetPublicModel(model.PublicModel).
		SetTemplateID(model.TemplateID).
		SetNillableUpstreamModelID(optionalStringPtr(model.UpstreamModelID)).
		SetRequestShape(model.RequestShape).
		SetStatus(model.Status).
		SetCapabilities(normalizeJSONMap(model.Capabilities)).
		SetDefaults(normalizeJSONMap(model.Defaults)).
		SetLimits(normalizeJSONMap(model.Limits)).
		SetSupportedOptions(normalizeJSONMap(model.SupportedOptions)).
		SetExtraBodyAllow(model.ExtraBodyAllow).
		SetSortOrder(model.SortOrder)
	if model.DisplayName != nil {
		builder.SetDisplayName(*model.DisplayName)
	}
	entModel, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create video model: %w", err)
	}
	model.ID = entModel.ID
	return nil
}

func (r *videoRepository) UpdateModel(ctx context.Context, model *service.VideoModel) error {
	builder := r.client.VideoModel.UpdateOneID(model.ID).
		SetPublicModel(model.PublicModel).
		SetTemplateID(model.TemplateID).
		SetRequestShape(model.RequestShape).
		SetStatus(model.Status).
		SetCapabilities(normalizeJSONMap(model.Capabilities)).
		SetDefaults(normalizeJSONMap(model.Defaults)).
		SetLimits(normalizeJSONMap(model.Limits)).
		SetSupportedOptions(normalizeJSONMap(model.SupportedOptions)).
		SetExtraBodyAllow(model.ExtraBodyAllow).
		SetSortOrder(model.SortOrder)
	if strings.TrimSpace(model.UpstreamModelID) != "" {
		builder.SetUpstreamModelID(model.UpstreamModelID)
	} else {
		builder.ClearUpstreamModelID()
	}
	if model.DisplayName != nil {
		builder.SetDisplayName(*model.DisplayName)
	} else {
		builder.ClearDisplayName()
	}
	if _, err := builder.Save(ctx); err != nil {
		return fmt.Errorf("update video model: %w", err)
	}
	return nil
}

func (r *videoRepository) DeleteModel(ctx context.Context, id int64) error {
	if err := r.client.VideoModel.DeleteOneID(id).Exec(ctx); err != nil {
		if dbent.IsNotFound(err) {
			return service.ErrVideoModelNotFound
		}
		return fmt.Errorf("delete video model: %w", err)
	}
	return nil
}

func (r *videoRepository) CreateTask(ctx context.Context, task *service.VideoGenerationTask) error {
	builder := r.client.VideoGenerationTask.Create().
		SetPublicID(task.PublicID).
		SetUserID(task.UserID).
		SetAPIKeyID(task.APIKeyID).
		SetAccountID(task.AccountID).
		SetVideoModelID(task.VideoModelID).
		SetRequestedModel(task.RequestedModel).
		SetUpstreamModel(task.UpstreamModel).
		SetStatus(task.Status).
		SetProgress(task.Progress).
		SetBillingState(task.BillingState).
		SetRequestPayload(normalizeJSONMap(task.RequestPayload)).
		SetBillingMode(task.BillingMode).
		SetUnitPrice(task.UnitPrice).
		SetReservedCost(task.ReservedCost).
		SetEstimatedCost(task.EstimatedCost).
		SetActualCost(task.ActualCost).
		SetPollCount(task.PollCount)

	if task.GroupID != nil {
		builder.SetGroupID(*task.GroupID)
	}
	if task.ChannelID != nil {
		builder.SetChannelID(*task.ChannelID)
	}
	if task.UpstreamTaskID != nil {
		builder.SetUpstreamTaskID(*task.UpstreamTaskID)
	}
	if task.UpstreamRequestPayload != nil {
		builder.SetUpstreamRequestPayload(normalizeJSONMap(task.UpstreamRequestPayload))
	}
	if task.UpstreamResponsePayload != nil {
		builder.SetUpstreamResponsePayload(normalizeJSONMap(task.UpstreamResponsePayload))
	}
	if task.ResultPayload != nil {
		builder.SetResultPayload(normalizeJSONMap(task.ResultPayload))
	}
	if task.ErrorCode != nil {
		builder.SetErrorCode(*task.ErrorCode)
	}
	if task.ErrorMessage != nil {
		builder.SetErrorMessage(*task.ErrorMessage)
	}
	if task.ContentURL != nil {
		builder.SetContentURL(*task.ContentURL)
	}
	if task.UpstreamContentURL != nil {
		builder.SetUpstreamContentURL(*task.UpstreamContentURL)
	}
	if task.LocalContentURL != nil {
		builder.SetLocalContentURL(*task.LocalContentURL)
	}
	if task.UnitSeconds != nil {
		builder.SetUnitSeconds(*task.UnitSeconds)
	}
	if task.RequestedSeconds != nil {
		builder.SetRequestedSeconds(*task.RequestedSeconds)
	}
	if task.BillableSeconds != nil {
		builder.SetBillableSeconds(*task.BillableSeconds)
	}
	if task.IdempotencyKey != nil {
		builder.SetIdempotencyKey(*task.IdempotencyKey)
	}
	if task.SubmittedAt != nil {
		builder.SetSubmittedAt(*task.SubmittedAt)
	}
	if task.StartedAt != nil {
		builder.SetStartedAt(*task.StartedAt)
	}
	if task.CompletedAt != nil {
		builder.SetCompletedAt(*task.CompletedAt)
	}
	if task.ExpiresAt != nil {
		builder.SetExpiresAt(*task.ExpiresAt)
	}
	if task.NextPollAt != nil {
		builder.SetNextPollAt(*task.NextPollAt)
	}
	if task.LockedUntil != nil {
		builder.SetLockedUntil(*task.LockedUntil)
	}
	entTask, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create video task: %w", err)
	}
	task.ID = entTask.ID
	task.CreatedAt = entTask.CreatedAt
	task.UpdatedAt = entTask.UpdatedAt
	return nil
}

func (r *videoRepository) GetTaskByPublicID(ctx context.Context, publicID string) (*service.VideoGenerationTask, error) {
	entTask, err := r.client.VideoGenerationTask.Query().
		Where(dbvideogenerationtask.PublicIDEQ(strings.TrimSpace(publicID))).
		WithVideoModel(func(q *dbent.VideoModelQuery) { q.WithTemplate() }).
		Only(ctx)
	if dbent.IsNotFound(err) {
		return nil, service.ErrVideoTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get video task: %w", err)
	}
	return videoTaskToService(entTask), nil
}

func (r *videoRepository) GetTaskByIdempotencyKey(ctx context.Context, key string) (*service.VideoGenerationTask, error) {
	entTask, err := r.client.VideoGenerationTask.Query().
		Where(dbvideogenerationtask.IdempotencyKeyEQ(strings.TrimSpace(key))).
		WithVideoModel(func(q *dbent.VideoModelQuery) { q.WithTemplate() }).
		Only(ctx)
	if dbent.IsNotFound(err) {
		return nil, service.ErrVideoTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get video task by idempotency key: %w", err)
	}
	return videoTaskToService(entTask), nil
}

func (r *videoRepository) ListTasksByAPIKey(ctx context.Context, apiKeyID int64, limit int, after string) ([]service.VideoGenerationTask, error) {
	q := r.client.VideoGenerationTask.Query().
		Where(dbvideogenerationtask.APIKeyIDEQ(apiKeyID)).
		WithVideoModel(func(q *dbent.VideoModelQuery) { q.WithTemplate() }).
		Order(dbent.Desc(dbvideogenerationtask.FieldCreatedAt), dbent.Desc(dbvideogenerationtask.FieldID)).
		Limit(limit)
	if after != "" {
		afterTask, err := r.client.VideoGenerationTask.Query().
			Where(dbvideogenerationtask.PublicIDEQ(after)).
			Only(ctx)
		if err == nil {
			q = q.Where(dbvideogenerationtask.CreatedAtLT(afterTask.CreatedAt))
		}
	}
	entTasks, err := q.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list video tasks: %w", err)
	}
	out := make([]service.VideoGenerationTask, 0, len(entTasks))
	for _, entTask := range entTasks {
		out = append(out, *videoTaskToService(entTask))
	}
	return out, nil
}

func (r *videoRepository) ListTasks(ctx context.Context, filter service.VideoTaskFilter) ([]service.VideoGenerationTask, int64, error) {
	q := r.client.VideoGenerationTask.Query().
		WithVideoModel(func(q *dbent.VideoModelQuery) { q.WithTemplate() })
	if filter.Status != "" {
		q = q.Where(dbvideogenerationtask.StatusEQ(strings.TrimSpace(filter.Status)))
	}
	if filter.Model != "" {
		model := strings.TrimSpace(filter.Model)
		q = q.Where(dbvideogenerationtask.Or(
			dbvideogenerationtask.RequestedModelContainsFold(model),
			dbvideogenerationtask.UpstreamModelContainsFold(model),
		))
	}
	if filter.UserID > 0 {
		q = q.Where(dbvideogenerationtask.UserIDEQ(filter.UserID))
	}
	if filter.APIKeyID > 0 {
		q = q.Where(dbvideogenerationtask.APIKeyIDEQ(filter.APIKeyID))
	}
	if filter.StartAt != nil {
		q = q.Where(dbvideogenerationtask.CreatedAtGTE(*filter.StartAt))
	}
	if filter.EndAt != nil {
		q = q.Where(dbvideogenerationtask.CreatedAtLTE(*filter.EndAt))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count video tasks: %w", err)
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	entTasks, err := q.Order(dbent.Desc(dbvideogenerationtask.FieldCreatedAt), dbent.Desc(dbvideogenerationtask.FieldID)).
		Offset(filter.Offset).
		Limit(filter.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list video tasks: %w", err)
	}
	out := make([]service.VideoGenerationTask, 0, len(entTasks))
	for _, entTask := range entTasks {
		out = append(out, *videoTaskToService(entTask))
	}
	return out, int64(total), nil
}

func (r *videoRepository) ClaimDueTasks(ctx context.Context, limit int, lockFor time.Duration) ([]service.VideoGenerationTask, error) {
	if limit <= 0 {
		limit = 10
	}
	if lockFor <= 0 {
		lockFor = 2 * time.Minute
	}
	now := time.Now().UTC()
	rows, err := r.db.QueryContext(ctx, `
		WITH due AS (
			SELECT id
			FROM video_generation_tasks
			WHERE status IN ('queued', 'in_progress')
			  AND (next_poll_at IS NULL OR next_poll_at <= $1)
			  AND (locked_until IS NULL OR locked_until <= $1)
			ORDER BY COALESCE(next_poll_at, created_at), id
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		UPDATE video_generation_tasks t
		SET locked_until = $3, updated_at = $1
		FROM due
		WHERE t.id = due.id
		RETURNING t.id
	`, now, limit, now.Add(lockFor))
	if err != nil {
		return nil, fmt.Errorf("claim video tasks: %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0, limit)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan claimed video task: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate claimed video tasks: %w", err)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	entTasks, err := r.client.VideoGenerationTask.Query().
		Where(dbvideogenerationtask.IDIn(ids...)).
		WithVideoModel(func(q *dbent.VideoModelQuery) { q.WithTemplate() }).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("load claimed video tasks: %w", err)
	}
	out := make([]service.VideoGenerationTask, 0, len(entTasks))
	for _, entTask := range entTasks {
		out = append(out, *videoTaskToService(entTask))
	}
	return out, nil
}

func (r *videoRepository) MarkTaskSubmitted(ctx context.Context, publicID, upstreamTaskID string, upstreamRequest, upstreamResponse map[string]any, nextPollAt time.Time) error {
	_, err := r.client.VideoGenerationTask.Update().
		Where(dbvideogenerationtask.PublicIDEQ(publicID)).
		SetUpstreamTaskID(upstreamTaskID).
		SetStatus(service.VideoStatusInProgress).
		SetProgress(1).
		SetSubmittedAt(time.Now().UTC()).
		SetStartedAt(time.Now().UTC()).
		SetNextPollAt(nextPollAt).
		ClearLockedUntil().
		SetUpstreamRequestPayload(normalizeJSONMap(upstreamRequest)).
		SetUpstreamResponsePayload(normalizeJSONMap(upstreamResponse)).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("mark video task submitted: %w", err)
	}
	return nil
}

func (r *videoRepository) MarkTaskPollResult(ctx context.Context, publicID string, status string, progress int, upstreamResponse map[string]any, contentURL *string, resultPayload map[string]any, nextPollAt *time.Time) error {
	builder := r.client.VideoGenerationTask.Update().
		Where(dbvideogenerationtask.PublicIDEQ(publicID)).
		SetStatus(status).
		SetProgress(progress).
		AddPollCount(1).
		ClearLockedUntil().
		SetUpstreamResponsePayload(normalizeJSONMap(upstreamResponse)).
		SetResultPayload(normalizeJSONMap(resultPayload))
	if contentURL != nil {
		builder.SetUpstreamContentURL(*contentURL)
		builder.SetContentURL(*contentURL)
	}
	if nextPollAt != nil {
		builder.SetNextPollAt(*nextPollAt)
	} else {
		builder.ClearNextPollAt()
	}
	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("mark video task poll result: %w", err)
	}
	return nil
}

func (r *videoRepository) MarkTaskFailed(ctx context.Context, publicID, code, message string) error {
	err := r.settleTerminalTask(ctx, publicID, service.VideoStatusFailed, 0, code, message, 0, nil, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("mark video task failed: %w", err)
	}
	return nil
}

func (r *videoRepository) MarkTaskFailedNoSettlement(ctx context.Context, publicID, code, message string) error {
	_, err := r.client.VideoGenerationTask.Update().
		Where(dbvideogenerationtask.PublicIDEQ(publicID)).
		SetStatus(service.VideoStatusFailed).
		SetProgress(0).
		SetErrorCode(code).
		SetErrorMessage(message).
		SetCompletedAt(time.Now().UTC()).
		ClearNextPollAt().
		ClearLockedUntil().
		Save(ctx)
	if err != nil {
		return fmt.Errorf("mark video task failed without settlement: %w", err)
	}
	return nil
}

func (r *videoRepository) MarkTaskCompleted(ctx context.Context, publicID string, progress int, contentURL *string, resultPayload map[string]any, billableSeconds *int, actualCost float64, expiresAt time.Time) error {
	err := r.settleTerminalTask(ctx, publicID, service.VideoStatusCompleted, progress, "", "", actualCost, contentURL, resultPayload, billableSeconds, &expiresAt)
	if err != nil {
		return fmt.Errorf("mark video task completed: %w", err)
	}
	return nil
}

func (r *videoRepository) MarkTaskCancelled(ctx context.Context, publicID string) (bool, error) {
	affected, err := r.settleCancelledTask(ctx, publicID)
	if err != nil {
		return false, fmt.Errorf("mark video task cancelled: %w", err)
	}
	return affected > 0, nil
}

func (r *videoRepository) MarkTaskExpired(ctx context.Context, publicID string) error {
	_, err := r.client.VideoGenerationTask.Update().
		Where(
			dbvideogenerationtask.PublicIDEQ(publicID),
			dbvideogenerationtask.StatusEQ(service.VideoStatusCompleted),
			dbvideogenerationtask.ExpiresAtLTE(time.Now().UTC()),
		).
		SetStatus(service.VideoStatusExpired).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("mark video task expired: %w", err)
	}
	return nil
}

func (r *videoRepository) RequeueTask(ctx context.Context, publicID string, nextPollAt time.Time) error {
	_, err := r.client.VideoGenerationTask.Update().
		Where(dbvideogenerationtask.PublicIDEQ(publicID)).
		SetStatus(service.VideoStatusQueued).
		SetProgress(0).
		SetNextPollAt(nextPollAt).
		SetPollCount(0).
		ClearErrorCode().
		ClearErrorMessage().
		ClearLockedUntil().
		Save(ctx)
	if err != nil {
		return fmt.Errorf("requeue video task: %w", err)
	}
	return nil
}

func (r *videoRepository) ScheduleTaskRetry(ctx context.Context, publicID string, nextPollAt time.Time, code, message string) error {
	_, err := r.client.VideoGenerationTask.Update().
		Where(dbvideogenerationtask.PublicIDEQ(publicID)).
		SetNextPollAt(nextPollAt).
		SetErrorCode(code).
		SetErrorMessage(message).
		AddPollCount(1).
		ClearLockedUntil().
		Save(ctx)
	if err != nil {
		return fmt.Errorf("schedule video task retry: %w", err)
	}
	return nil
}

func (r *videoRepository) settleTerminalTask(ctx context.Context, publicID, status string, progress int, code, message string, actualCost float64, contentURL *string, resultPayload map[string]any, billableSeconds *int, expiresAt *time.Time) error {
	if r.db == nil {
		return fmt.Errorf("sql db is required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	var taskID, userID int64
	var reservedCost float64
	var billingState string
	row := tx.QueryRowContext(ctx, `
		SELECT id, user_id, reserved_cost, billing_state
		FROM video_generation_tasks
		WHERE public_id = $1
		FOR UPDATE
	`, publicID)
	if err = row.Scan(&taskID, &userID, &reservedCost, &billingState); err != nil {
		return err
	}
	refund := 0.0
	if billingState == service.VideoBillingStateReserved {
		refund = reservedCost - actualCost
		if refund < 0 {
			refund = 0
		}
	}

	query := `
		UPDATE video_generation_tasks
		SET status = $2,
		    progress = $3,
		    error_code = NULLIF($4, ''),
		    error_message = NULLIF($5, ''),
		    actual_cost = $6,
		    billing_state = $7,
		    completed_at = $8,
		    next_poll_at = NULL,
		    locked_until = NULL,
		    updated_at = $8`
	args := []any{publicID, status, progress, code, message, actualCost, service.VideoBillingStateSettled, now}
	nextArg := 9
	if contentURL != nil {
		query += fmt.Sprintf(", content_url = $%d, upstream_content_url = $%d", nextArg, nextArg)
		args = append(args, *contentURL)
		nextArg++
	}
	if resultPayload != nil {
		payloadJSON, marshalErr := json.Marshal(normalizeJSONMap(resultPayload))
		if marshalErr != nil {
			err = marshalErr
			return err
		}
		query += fmt.Sprintf(", result_payload = $%d::jsonb", nextArg)
		args = append(args, string(payloadJSON))
		nextArg++
	}
	if billableSeconds != nil {
		query += fmt.Sprintf(", billable_seconds = $%d", nextArg)
		args = append(args, *billableSeconds)
		nextArg++
	}
	if status == service.VideoStatusCompleted {
		if expiresAt == nil || expiresAt.IsZero() {
			fallback := now.Add(24 * time.Hour)
			expiresAt = &fallback
		}
		query += fmt.Sprintf(", expires_at = $%d", nextArg)
		args = append(args, *expiresAt)
		nextArg++
	}
	query += " WHERE public_id = $1"
	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		return err
	}
	if refund > 0 {
		if _, err = tx.ExecContext(ctx, `UPDATE users SET balance = balance + $1, updated_at = $2 WHERE id = $3`, refund, now, userID); err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}

func (r *videoRepository) settleCancelledTask(ctx context.Context, publicID string) (int, error) {
	if r.db == nil {
		return 0, fmt.Errorf("sql db is required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	var userID int64
	var reservedCost float64
	var status, billingState string
	row := tx.QueryRowContext(ctx, `
		SELECT user_id, reserved_cost, status, billing_state
		FROM video_generation_tasks
		WHERE public_id = $1
		FOR UPDATE
	`, publicID)
	if err = row.Scan(&userID, &reservedCost, &status, &billingState); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	if status != service.VideoStatusQueued && status != service.VideoStatusInProgress {
		err = tx.Commit()
		return 0, err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE video_generation_tasks
		SET status = $2,
		    actual_cost = 0,
		    billing_state = $3,
		    completed_at = $4,
		    next_poll_at = NULL,
		    locked_until = NULL,
		    updated_at = $4
		WHERE public_id = $1
	`, publicID, service.VideoStatusCancelled, service.VideoBillingStateSettled, now); err != nil {
		return 0, err
	}
	if billingState == service.VideoBillingStateReserved && reservedCost > 0 {
		if _, err = tx.ExecContext(ctx, `UPDATE users SET balance = balance + $1, updated_at = $2 WHERE id = $3`, reservedCost, now, userID); err != nil {
			return 0, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func videoTemplateToService(t *dbent.VideoCallTemplate) *service.VideoCallTemplate {
	if t == nil {
		return nil
	}
	return &service.VideoCallTemplate{
		ID:            t.ID,
		Name:          t.Name,
		CreateMethod:  t.CreateMethod,
		CreatePath:    t.CreatePath,
		QueryMethod:   t.QueryMethod,
		QueryPath:     t.QueryPath,
		ContentMethod: t.ContentMethod,
		ContentPath:   t.ContentPath,
		CancelMethod:  t.CancelMethod,
		CancelPath:    t.CancelPath,
		StatusMapping: t.StatusMapping,
		ResultMapping: t.ResultMapping,
		ErrorMapping:  t.ErrorMapping,
		PollConfig:    t.PollConfig,
		TimeoutConfig: t.TimeoutConfig,
		Status:        t.Status,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

func videoModelToService(m *dbent.VideoModel) *service.VideoModel {
	if m == nil {
		return nil
	}
	out := &service.VideoModel{
		ID:               m.ID,
		PublicModel:      m.PublicModel,
		DisplayName:      m.DisplayName,
		TemplateID:       m.TemplateID,
		UpstreamModelID:  derefString(m.UpstreamModelID),
		RequestShape:     m.RequestShape,
		Status:           m.Status,
		Capabilities:     m.Capabilities,
		Defaults:         m.Defaults,
		Limits:           m.Limits,
		SupportedOptions: m.SupportedOptions,
		ExtraBodyAllow:   m.ExtraBodyAllow,
		SortOrder:        m.SortOrder,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
	if m.Edges.Template != nil {
		out.Template = videoTemplateToService(m.Edges.Template)
	}
	return out
}

func videoTaskToService(t *dbent.VideoGenerationTask) *service.VideoGenerationTask {
	if t == nil {
		return nil
	}
	out := &service.VideoGenerationTask{
		ID:                      t.ID,
		PublicID:                t.PublicID,
		UserID:                  t.UserID,
		APIKeyID:                t.APIKeyID,
		GroupID:                 t.GroupID,
		AccountID:               t.AccountID,
		ChannelID:               t.ChannelID,
		VideoModelID:            t.VideoModelID,
		RequestedModel:          t.RequestedModel,
		UpstreamModel:           t.UpstreamModel,
		UpstreamTaskID:          t.UpstreamTaskID,
		Status:                  t.Status,
		Progress:                t.Progress,
		BillingState:            t.BillingState,
		RequestPayload:          t.RequestPayload,
		UpstreamRequestPayload:  t.UpstreamRequestPayload,
		UpstreamResponsePayload: t.UpstreamResponsePayload,
		ResultPayload:           t.ResultPayload,
		ErrorCode:               t.ErrorCode,
		ErrorMessage:            t.ErrorMessage,
		ContentURL:              t.ContentURL,
		UpstreamContentURL:      t.UpstreamContentURL,
		LocalContentURL:         t.LocalContentURL,
		BillingMode:             t.BillingMode,
		UnitPrice:               t.UnitPrice,
		UnitSeconds:             t.UnitSeconds,
		RequestedSeconds:        t.RequestedSeconds,
		BillableSeconds:         t.BillableSeconds,
		ReservedCost:            t.ReservedCost,
		EstimatedCost:           t.EstimatedCost,
		ActualCost:              t.ActualCost,
		IdempotencyKey:          t.IdempotencyKey,
		SubmittedAt:             t.SubmittedAt,
		StartedAt:               t.StartedAt,
		CompletedAt:             t.CompletedAt,
		ExpiresAt:               t.ExpiresAt,
		NextPollAt:              t.NextPollAt,
		PollCount:               t.PollCount,
		LockedUntil:             t.LockedUntil,
		CreatedAt:               t.CreatedAt,
		UpdatedAt:               t.UpdatedAt,
	}
	if t.Edges.VideoModel != nil {
		out.VideoModel = videoModelToService(t.Edges.VideoModel)
	}
	return out
}

// optionalStringPtr returns nil for empty/whitespace strings so the column is
// stored as NULL, otherwise a pointer to the trimmed value.
func optionalStringPtr(s string) *string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
