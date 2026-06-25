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

// VideoCallTemplate describes the HTTP shape of an upstream async video API.
type VideoCallTemplate struct {
	ent.Schema
}

func (VideoCallTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "video_call_templates"},
	}
}

func (VideoCallTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (VideoCallTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(100).NotEmpty(),
		field.String("create_method").MaxLen(10).Default("POST"),
		field.String("create_path").MaxLen(500).NotEmpty(),
		field.String("query_method").MaxLen(10).Default("GET"),
		field.String("query_path").MaxLen(500).NotEmpty(),
		field.String("content_method").MaxLen(10).Optional().Nillable(),
		field.String("content_path").MaxLen(500).Optional().Nillable(),
		field.String("cancel_method").MaxLen(10).Optional().Nillable(),
		field.String("cancel_path").MaxLen(500).Optional().Nillable(),
		field.JSON("status_mapping", map[string]string{}).
			Default(map[string]string{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("result_mapping", map[string]string{}).
			Default(map[string]string{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("error_mapping", map[string]string{}).
			Default(map[string]string{}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("poll_config", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("timeout_config", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("status").MaxLen(20).Default("active"),
	}
}

func (VideoCallTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("video_models", VideoModel.Type),
	}
}

func (VideoCallTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique(),
		index.Fields("status"),
	}
}
