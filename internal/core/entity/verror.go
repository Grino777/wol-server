package entity

// Compile-time check.
var (
	_ error  = (*ValidationError)(nil)
	_ Public = (*ValidationError)(nil)
)

// ValidationError is a general error type that returns any input validation
// errors in application.
type ValidationError struct {
	Err         error
	FieldErrors []FieldErrorInfo
}

// FieldErrorInfo is a error information for the unprocessable field.
type FieldErrorInfo struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

func NewValidationError(err error, fields []FieldErrorInfo) error {
	return &ValidationError{
		Err:         err,
		FieldErrors: fields,
	}
}

// Error implements error interface.
func (e *ValidationError) Error() string {
	if e.Err != nil {
		return "validation error: " + e.Err.Error()
	}
	return "validation error"
}

// Unwrap implements error unwrapping interface
// and returns original error.
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// Public implements Public interface.
func (e *ValidationError) Public() interface{} {
	return map[string][]FieldErrorInfo{
		"errors": e.FieldErrors,
	}
}

// TODO: translate and error codes
// func fromValidationErrors(ctx context.Context, errs validator.ValidationErrors) (*ValidationError, error) {
// 	e := &ValidationError{
// 		Err:         errs,
// 		FieldErrors: make([]FieldErrorInfo, 0, len(errs)),
// 	}

// 	t, _ := validator.GetTranslator("en")

// 	for _, f := range errs {
// 		field := f.Namespace()

// 		// remove structure name from field path
// 		// example: UserDeviceNew.data[0].value -> data[0].value
// 		if i := strings.IndexRune(field, '.'); i > 0 {
// 			field = field[i+1:]
// 		}

// 		e.FieldErrors = append(e.FieldErrors, FieldErrorInfo{
// 			Code:    f.ActualTag(),
// 			Field:   field,
// 			Message: f.Translate(t),
// 		})
// 	}
// 	return e, nil
// }

// func newFromValidationErrors(ctx context.Context, errs validator.ValidationErrors, fieldName string) (*ValidationError, error) {
// 	e := &ValidationError{
// 		Err:         errs,
// 		FieldErrors: make([]FieldErrorInfo, 0, len(errs)),
// 	}

// 	t, _ := validator.GetTranslator("en")

// 	for _, f := range errs {
// 		field := f.Namespace()
// 		if field == "" {
// 			field = fieldName
// 		} else {
// 			// Remove the structure name from the field path, if present
// 			if i := strings.IndexRune(field, '.'); i > 0 {
// 				field = field[i+1:]
// 			}
// 		}

// 		e.FieldErrors = append(e.FieldErrors, FieldErrorInfo{
// 			Code:    f.ActualTag(),
// 			Field:   field,
// 			Message: f.Translate(t),
// 		})
// 	}
// 	return e, nil
// }

// func tryWrapAsValidationError(ctx context.Context, err error) error {
// 	var errs validator.ValidationErrors
// 	if errors.As(err, &errs) {
// 		vErr, er := fromValidationErrors(ctx, errs)
// 		if er != nil {
// 			return er
// 		}
// 		return vErr
// 	}
// 	return err
// }

// func newTryWrapAsValidationError(ctx context.Context, err error, fieldName string) error {
// 	var errs validator.ValidationErrors
// 	if errors.As(err, &errs) {
// 		vErr, er := newFromValidationErrors(ctx, errs, fieldName)
// 		if er != nil {
// 			return er
// 		}
// 		return vErr
// 	}
// 	return err
// }
