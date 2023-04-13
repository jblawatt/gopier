package core

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v3"
)


type TemplateContext struct {
	Values map[interface{}]interface{}
}

func interpolate(strOrTemplate string, ctx *TemplateContext) string {
	tpl, _ := template.New("").Parse(strOrTemplate)
	buff := new(bytes.Buffer)
	tpl.Execute(buff, ctx)
	processed := buff.String()
	log.Printf("interpolating string %s to %s\n", strOrTemplate, processed)
	return processed
}

func mergeMaps(a, b map[interface{}]interface{}) map[interface{}]interface{} {
    out := make(map[interface{}]interface{}, len(a))
    for k, v := range a {
        out[k] = v
    }
    for k, v := range b {
        // If you use map[string]interface{}, ok is always false here.
        // Because yaml.Unmarshal will give you map[interface{}]interface{}.
        if v, ok := v.(map[interface{}]interface{}); ok {
            if bv, ok := out[k]; ok {
                if bv, ok := bv.(map[interface{}]interface{}); ok {
                    out[k] = mergeMaps(bv, v)
                    continue
                }
            }
        }
        out[k] = v
    }
    return out
}


func CreateTemplateContext(defaultValuesPath string, valuesPath string) (*TemplateContext, error) {

  defaultValuesContent, derr := ioutil.ReadFile(defaultValuesPath)
  if derr != nil {
    return nil, formatError("Error reading default values: %w", derr)
  }

	defaultValues := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(defaultValuesContent, &defaultValues); err != nil {
    return nil, formatError("Error unmarshalling default values: %w", err)
  }

  valuesContent, verr := ioutil.ReadFile(valuesPath)
  if verr != nil {
    return nil, formatError("Error reading default values: %w", derr)
  }

	values := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(valuesContent, &values); err != nil {
    return nil, formatError("Error unmarshalling default values: %w", err)
  }

  out := mergeMaps(defaultValues, values)
  log.Println("------------------------------------------")
  log.Println(out)
  log.Println("------------------------------------------")

	return &TemplateContext{Values: out}, nil
}
