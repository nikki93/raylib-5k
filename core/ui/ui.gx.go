package ui

//gx:extern invalid
type uiResult func(...interface{}) uiResult

//gx:extern ui
func UI(tag string) uiResult
