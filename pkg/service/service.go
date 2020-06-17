package service

type Service interface {
	Exec(filename string, args... interface{}) (interface{}, error);
	GetName() string;
}
