package feature

import (
	"context"
	"net/http"
)

type contextKey string

const featureContextKey contextKey = "feature_context"

// Context holds attributes for flag evaluation.
type Context struct {
	// Key is the unique user/entity identifier.
	Key string

	// Name is the display name.
	Name string

	// Email is the email address.
	Email string

	// IP is the IP address.
	IP string

	// Country is the country code.
	Country string

	// Custom holds custom attributes.
	Custom map[string]interface{}

	// Anonymous indicates an anonymous user.
	Anonymous bool

	// Groups holds group memberships.
	Groups []string
}

// NewContext creates an evaluation context with the given key.
func NewContext(key string) *Context {
	return &Context{
		Key:    key,
		Custom: make(map[string]interface{}),
		Groups: []string{},
	}
}

// WithAttribute adds a custom attribute to the context.
func (c *Context) WithAttribute(key string, value interface{}) *Context {
	if c.Custom == nil {
		c.Custom = make(map[string]interface{})
	}
	c.Custom[key] = value
	return c
}

// WithName sets the display name.
func (c *Context) WithName(name string) *Context {
	c.Name = name
	return c
}

// WithEmail sets the email address.
func (c *Context) WithEmail(email string) *Context {
	c.Email = email
	return c
}

// WithIP sets the IP address.
func (c *Context) WithIP(ip string) *Context {
	c.IP = ip
	return c
}

// WithCountry sets the country code.
func (c *Context) WithCountry(country string) *Context {
	c.Country = country
	return c
}

// WithAnonymous sets the anonymous flag.
func (c *Context) WithAnonymous(anonymous bool) *Context {
	c.Anonymous = anonymous
	return c
}

// WithGroup adds a group membership.
func (c *Context) WithGroup(group string) *Context {
	c.Groups = append(c.Groups, group)
	return c
}

// WithGroups sets all group memberships.
func (c *Context) WithGroups(groups []string) *Context {
	c.Groups = groups
	return c
}

// getAttribute retrieves an attribute value by name.
func (c *Context) getAttribute(name string) interface{} {
	if c == nil {
		return nil
	}

	switch name {
	case "key":
		return c.Key
	case "name":
		return c.Name
	case "email":
		return c.Email
	case "ip":
		return c.IP
	case "country":
		return c.Country
	case "anonymous":
		return c.Anonymous
	case "groups":
		return c.Groups
	default:
		if c.Custom != nil {
			return c.Custom[name]
		}
		return nil
	}
}

// WithContext adds the feature context to a standard context.Context.
func WithContext(ctx context.Context, fctx *Context) context.Context {
	return context.WithValue(ctx, featureContextKey, fctx)
}

// FromContext extracts the feature context from a standard context.Context.
func FromContext(ctx context.Context) *Context {
	if ctx == nil {
		return nil
	}
	fctx, _ := ctx.Value(featureContextKey).(*Context)
	return fctx
}

// ContextFromHTTPRequest extracts context from an HTTP request.
func ContextFromHTTPRequest(r *http.Request) *Context {
	if r == nil {
		return NewContext("")
	}

	ctx := NewContext("")

	// Try to extract IP from common headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		ctx.IP = ip
	} else if ip := r.Header.Get("X-Real-IP"); ip != "" {
		ctx.IP = ip
	} else {
		ctx.IP = r.RemoteAddr
	}

	// Extract country from common header
	if country := r.Header.Get("CF-IPCountry"); country != "" {
		ctx.Country = country
	}

	// Check for user-id header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		ctx.Key = userID
	}

	return ctx
}
