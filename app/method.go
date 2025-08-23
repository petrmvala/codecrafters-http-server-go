package main

const (
	methodGet  = "GET"
	methodPost = "POST"
)

var knownMethods = map[string]bool{
	methodGet:  true,
	methodPost: true,
}
