package crud

type (
	StorageInput interface {
		// GetArgs returns the arguments to be passed to the query.
		// The arguments should be in the same order as the query's placeholders.
		GetArgs() []interface{}
	}
	StorageOutput interface {
		GetInstance() StorageOutput
	}
)
