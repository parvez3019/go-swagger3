module github.com/parvez3019/go-swagger3/examples/chi

go 1.26.0

require (
	github.com/go-chi/chi/v5 v5.5.0
	github.com/parvez3019/go-swagger3 v1.2.0
	github.com/parvez3019/go-swagger3/swagger/chi v0.0.0
)

replace (
	github.com/parvez3019/go-swagger3 => ../..
	github.com/parvez3019/go-swagger3/swagger/chi => ../../swagger/chi
)
