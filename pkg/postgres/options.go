package postgres

// Option configures the Client.
type Option func(*Client)

// WithLogger sets a custom logger for the client.
func WithLogger(logger Logger) Option {
	return func(c *Client) {
		if logger != nil {
			c.logger = logger
		}
	}
}

// QueryHook is called before and after query execution.
type QueryHook interface {
	BeforeQuery(sql string, args []any)
	AfterQuery(sql string, args []any, err error)
}

// WithQueryHook sets a query hook for observability.
func WithQueryHook(hook QueryHook) Option {
	return func(c *Client) {
		c.queryHook = hook
	}
}
