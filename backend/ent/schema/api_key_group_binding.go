package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// APIKeyGroupBinding binds an API key to a group for multi-group routing,
// carrying priority/weight/enabled for the group selection algorithm.
type APIKeyGroupBinding struct {
	ent.Schema
}

func (APIKeyGroupBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "api_key_group_bindings"},
	}
}

func (APIKeyGroupBinding) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (APIKeyGroupBinding) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("api_key_id"),
		field.Int64("group_id"),
		field.Int("priority").
			Default(0).
			Comment("Lower value = higher priority"),
		field.Int("weight").
			Default(100).
			Comment("Weight within the same priority bucket (>0)"),
		field.Bool("enabled").
			Default(true),
	}
}

func (APIKeyGroupBinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("api_key", APIKey.Type).
			Ref("group_bindings").
			Field("api_key_id").
			Required().
			Unique(),
		edge.From("group", Group.Type).
			Ref("api_key_group_bindings").
			Field("group_id").
			Required().
			Unique(),
	}
}

func (APIKeyGroupBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("api_key_id", "group_id").Unique(),
		index.Fields("api_key_id"),
		index.Fields("group_id"),
	}
}
