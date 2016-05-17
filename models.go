package main

type Thing struct {
	ID        string `json:"id"`
	APIURL    string `json:"apiUrl"` // self ?
	PrefLabel string `json:"prefLabel,omitempty"`
}

type People struct {
    People  []Thing `json:"people"`
}
