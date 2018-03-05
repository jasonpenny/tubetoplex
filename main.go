package main

func main() {
	vi := downloadURL("https://www.youtube.com/watch?v=C0DPdy98e4c", 3, 2)

	err := createNFOFile(vi)
	if err != nil {
		panic(err)
	}
}
