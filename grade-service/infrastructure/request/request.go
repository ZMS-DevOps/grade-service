package request

type Validator interface {
	AreValidRequestData() error
}
