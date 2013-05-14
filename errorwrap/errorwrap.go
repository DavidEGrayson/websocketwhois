package errorwrap

type errorWrap struct {
  Message string
  InnerError error
}

func New(message string, err error) error {
  return &errorWrap{message, err}
}

func (e *errorWrap) Error() string {
  return e.Message + "  " + e.InnerError.Error()
}