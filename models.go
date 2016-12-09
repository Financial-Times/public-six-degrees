package main

type Thing struct {
	ID        string `json:"id"`
	APIURL    string `json:"apiUrl,omitempty"`
	PrefLabel string `json:"prefLabel,omitempty"`
}

type Content struct {
	ID     string `json:"id"`
	APIURL string `json:"apiUrl,omitempty"`
	Title  string `json:"title"`
}

type ConnectedPerson struct {
	Person  Thing     `json:"person"`
	Count   int       `json:"count"`
	Content []Content `json:"content"`
}
