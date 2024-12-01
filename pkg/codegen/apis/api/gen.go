package api

//go:generate oapi-codegen -generate types -package api -templates ../../templates -o types.go -import-mapping=../common/common.yaml:github.com/jacobbrewer1/f1-data/pkg/codegen/apis/common ./routes.yaml
//go:generate oapi-codegen -generate gorilla -package api -templates ../../templates -o server.go -import-mapping=../common/common.yaml:github.com/jacobbrewer1/f1-data/pkg/codegen/apis/common ./routes.yaml
//go:generate oapi-codegen -generate client -package api -templates ../../templates -o client.go -import-mapping=../common/common.yaml:github.com/jacobbrewer1/f1-data/pkg/codegen/apis/common ./routes.yaml
