package main

type Thing struct {
	ID        string `json:"id"`
	APIURL    string `json:"apiUrl,omitempty"` // self ?
	PrefLabel string `json:"prefLabel,omitempty"`
}

type ConnectedPerson struct {
	Person Thing `json:"person"`
	Count  int   `json:"count"`
}
