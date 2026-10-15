package dto

type MCPExportTarget struct {
	Kind string `json:"kind" description:"Kind of the exported item." enum:"workspace,collection,folder,request,environment,globals" validated_as:"item_kind"`
	Id   string `json:"id" description:"Id of the exported item." validated_as:"item_id"`
}
