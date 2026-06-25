package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// VideoGenerationTask stores async video generation lifecycle state.
type VideoGenerationTask struct {
	ent.Schema
}

func (VideoGenerationTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "video_generation_tasks"},
	}
}

func (VideoGenerationTask) Fields() []ent.Field {
	return []ent.Field{
		field.String("public_id").MaxLen(80).NotEmpty(),
		field.Int64("user_id"),
		field.Int64("api_key_id"),
		field.Int64("group_id").Optional().Nillable(),
		field.Int64("account_id"),
		field.Int64("channel_id").Optional().Nillable(),
		field.Int64("video_model_id"),
		field.String("requested_model").MaxLen(150).NotEmpty(),
		field.String("upstream_model").MaxLen(150).NotEmpty(),
		field.String("upstream_task_id").MaxLen(200).Optional().Nillable(),
		field.String("status").MaxLen(20).Default("queued"),
		field.Int("progress").Default(0),
		field.String("billing_state").MaxLen(20).Default("reserved"),
		field.JSON("request_payload", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("upstream_request_payload", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("upstream_response_payload", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("result_payload", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("error_code").MaxLen(100).Optional().Nillable(),
		field.String("error_message").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("content_url").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("upstream_content_url").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("local_content_url").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("billing_mode").MaxLen(20).NotEmpty(),
		field.Float("unit_price").SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).Default(0),
		field.Float("unit_seconds").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.Int("requested_seconds").Optional().Nillable(),
		field.Int("billable_seconds").Optional().Nillable(),
		field.Float("reserved_cost").SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).Default(0),
		field.Float("estimated_cost").SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).Default(0),
		field.Float("actual_cost").SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}).Default(0),
		field.String("idempotency_key").MaxLen(128).Optional().Nillable(),
		field.Time("submitted_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("started_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("completed_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("next_poll_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Int("poll_count").Default(0),
		field.Time("locked_until").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (VideoGenerationTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("video_generation_tasks").Field("user_id").Required().Unique(),
		edge.From("api_key", APIKey.Type).Ref("video_generation_tasks").Field("api_key_id").Required().Unique(),
		edge.From("account", Account.Type).Ref("video_generation_tasks").Field("account_id").Required().Unique(),
		edge.From("group", Group.Type).Ref("video_generation_tasks").Field("group_id").Unique(),
		edge.From("video_model", VideoModel.Type).Ref("tasks").Field("video_model_id").Required().Unique(),
	}
}

func (VideoGenerationTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("public_id").Unique(),
		index.Fields("status", "next_poll_at"),
		index.Fields("user_id", "created_at"),
		index.Fields("api_key_id", "created_at"),
		index.Fields("billing_state"),
		index.Fields("video_model_id"),
		index.Fields("upstream_task_id"),
		index.Fields("idempotency_key").Unique(),
	}
}
