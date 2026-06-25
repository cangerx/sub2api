package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// VideoModel is the public video model catalog entry.
type VideoModel struct {
	ent.Schema
}

func (VideoModel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "video_models"},
	}
}

func (VideoModel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (VideoModel) Fields() []ent.Field {
	return []ent.Field{
		field.String("public_model").MaxLen(100).NotEmpty(),
		field.String("display_name").MaxLen(100).Optional().Nillable(),
		field.Int64("template_id"),
		field.String("upstream_model_id").MaxLen(150).Optional().Nillable(),
		field.String("request_shape").MaxLen(50).NotEmpty(),
		field.String("status").MaxLen(20).Default("active"),
		field.JSON("capabilities", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("defaults", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("limits", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("supported_options", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("extra_body_allow", []string{}).
			Default([]string{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Int("sort_order").Default(0),
	}
}

func (VideoModel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("template", VideoCallTemplate.Type).
			Ref("video_models").
			Field("template_id").
			Required().
			Unique(),
		edge.To("tasks", VideoGenerationTask.Type),
	}
}

func (VideoModel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("public_model").Unique(),
		index.Fields("template_id"),
		index.Fields("status"),
		index.Fields("sort_order"),
	}
}
