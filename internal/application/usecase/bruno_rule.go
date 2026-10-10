package usecase

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/brunoruleform"
)

type brunoRule struct {
	replacement string
	form        enums.BrunoRuleForm
	arity       int
}

var brunoRules = map[string]brunoRule{
	"bru.getEnvVar":        {replacement: "pm.environment.get", form: brunoruleform.RenamedCall},
	"bru.setEnvVar":        {replacement: "pm.environment.set", form: brunoruleform.RenamedCall},
	"bru.getVar":           {replacement: "pm.variables.get", form: brunoruleform.RenamedCall},
	"bru.setVar":           {replacement: "pm.variables.set", form: brunoruleform.RenamedCall},
	"bru.getRequestVar":    {replacement: "pm.variables.get", form: brunoruleform.RenamedCall},
	"bru.getCollectionVar": {replacement: "pm.collectionVariables.get", form: brunoruleform.RenamedCall},
	"bru.getGlobalEnvVar":  {replacement: "pm.globals.get", form: brunoruleform.RenamedCall},
	"bru.setGlobalEnvVar":  {replacement: "pm.globals.set", form: brunoruleform.RenamedCall},
	"bru.getEnvName":       {replacement: "pm.environment.name", form: brunoruleform.EmptyCall},
	"bru.interpolate":      {replacement: "pm.variables.replaceIn", form: brunoruleform.RenamedCall},
	"req.getUrl":           {replacement: "pm.request.url.toString()", form: brunoruleform.EmptyCall},
	"req.url":              {replacement: "pm.request.url.toString()", form: brunoruleform.Property},
	"req.getMethod":        {replacement: "pm.request.method", form: brunoruleform.EmptyCall},
	"req.method":           {replacement: "pm.request.method", form: brunoruleform.Property},
	"req.getHeader":        {replacement: "pm.request.headers.get", form: brunoruleform.RenamedCall},
	"req.setHeader":        {replacement: "pm.request.headers.upsert({ key: %s, value: %s })", form: brunoruleform.WrappedCall, arity: 2},
	"req.getHeaders":       {replacement: "pm.request.headers.toObject()", form: brunoruleform.EmptyCall},
	"req.headers":          {replacement: "pm.request.headers.toObject()", form: brunoruleform.Property},
	"req.getBody":          {replacement: "JSON.parse(pm.request.body.raw)", form: brunoruleform.EmptyCall},
	"req.body":             {replacement: "JSON.parse(pm.request.body.raw)", form: brunoruleform.Property},
	"req.setBody":          {replacement: "pm.request.body.update(JSON.stringify(%s))", form: brunoruleform.WrappedCall, arity: 1},
	"res.getStatus":        {replacement: "pm.response.code", form: brunoruleform.EmptyCall},
	"res.status":           {replacement: "pm.response.code", form: brunoruleform.Property},
	"res.getBody":          {replacement: "pm.response.json()", form: brunoruleform.EmptyCall},
	"res.body":             {replacement: "pm.response.json()", form: brunoruleform.Property},
	"res.getHeader":        {replacement: "pm.response.headers.get", form: brunoruleform.RenamedCall},
	"res.getHeaders":       {replacement: "pm.response.headers.toObject()", form: brunoruleform.EmptyCall},
	"res.headers":          {replacement: "pm.response.headers.toObject()", form: brunoruleform.Property},
	"res.getResponseTime":  {replacement: "pm.response.responseTime", form: brunoruleform.EmptyCall},
	"res.responseTime":     {replacement: "pm.response.responseTime", form: brunoruleform.Property},
	"test":                 {replacement: "pm.test", form: brunoruleform.RenamedCall},
	"expect":               {replacement: "pm.expect", form: brunoruleform.RenamedCall},
}
