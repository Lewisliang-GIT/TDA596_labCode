package unit

import (
	"log"
)

func main() {
	ip := "0.0.0.0"
	port := "1234"
	address := ip + ":" + port
	id := hashString(address)
	//if (*idOverwrite != "") && (len(*idOverwrite) == 48) {
	//	id = *idOverwrite
	//}

	log.Printf("<LocalNode>: %+v\n")

	node := Node{Id: id, Address: NodeAddress(address)}
	node.creatChord()
}
