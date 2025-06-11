package resolvers

import (
	"context"

	"github.com/99designs/gqlgen/graphql/introspection"
	"github.com/DimaJoyti/go-coffee/accounts-service/internal/graphql/generated"
)

// ResolverWrapper wraps the main resolver to implement the generated interfaces
type ResolverWrapper struct {
	*Resolver
}

// NewResolverWrapper creates a new resolver wrapper
func NewResolverWrapper(resolver *Resolver) generated.ResolverRoot {
	return &ResolverWrapper{Resolver: resolver}
}

// Mutation returns the mutation resolver
func (w *ResolverWrapper) Mutation() generated.MutationResolver {
	return w.Resolver
}

// Query returns the query resolver
func (w *ResolverWrapper) Query() generated.QueryResolver {
	return w.Resolver
}

// __InputValue returns the input value resolver
func (w *ResolverWrapper) __InputValue() interface{} {
	return &inputValueResolver{w.Resolver}
}

// __Type returns the type resolver
func (w *ResolverWrapper) __Type() interface{} {
	return &typeResolver{w.Resolver}
}

// inputValueResolver implements the __InputValueResolver interface
type inputValueResolver struct {
	*Resolver
}

// IsDeprecated is the resolver for the isDeprecated field.
func (r *inputValueResolver) IsDeprecated(ctx context.Context, obj *introspection.InputValue) (bool, error) {
	return false, nil // Default implementation
}

// DeprecationReason is the resolver for the deprecationReason field.
func (r *inputValueResolver) DeprecationReason(ctx context.Context, obj *introspection.InputValue) (*string, error) {
	return nil, nil // Default implementation
}

// typeResolver implements the __TypeResolver interface
type typeResolver struct {
	*Resolver
}

// IsOneOf is the resolver for the isOneOf field.
func (r *typeResolver) IsOneOf(ctx context.Context, obj *introspection.Type) (*bool, error) {
	result := false
	return &result, nil // Default implementation
}
