package mgr

//ManagerMethod Define a type signature for all the manager methods
type ManagerMethod func([]string) int

//Null function
func Null([]string) int {
	return 1
}
