// hash functions
package chord

import (
	"crypto/sha1"
	"io"
	"math/big"
)

// also known as sha1.Size*8
const keySize = 160

// 2^160
var hashMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(keySize), nil)

func pow_2(x int) *big.Int {
	return new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(x)), nil)
}

func Hash(in string) *big.Int {
	hash1 := sha1.New()
	io.WriteString(hash1, in)

	num := new(big.Int).SetBytes(hash1.Sum(nil))
	//slice := hash1.Sum(nil)
	// num := new(big.Int).SetBytes(slice)
	//also write like:
	// num := new(big.Int)
	// num.SetBytes(slice)
	return num
}

// also func jump
// 0-base
// initalize start's fingertable, into No.fingerNum(2^(m))
func fingerEntry(startId *big.Int, fingerentryNum int) *big.Int {
	exponent := big.NewInt(int64(fingerentryNum))
	distance := new(big.Int).Exp(big.NewInt(2), exponent, nil)

	fingerid := startId
	fingerid.Add(fingerid, distance)
	fingerid.Mod(fingerid, hashMod)
	return fingerid
}

func Between(id *big.Int, left *big.Int, right *big.Int, isInclusive bool) bool {
	if isInclusive {
		return InclusiveBetween(id, left, right)
	} else {
		return ExclusiveBetween(id, left, right)
	}
}

// judge if inclusive id belongs to (left, right]
// as a circle, we know if left < right the interval dont contain 0
// otherwise, if left == right, it is also ok
func InclusiveBetween(id *big.Int, left *big.Int, right *big.Int) bool {
	if right.Cmp(left) == 1 {
		//right > left
		return id.Cmp(left) == 1 && right.Cmp(id) >= 0
	} else {
		// fmt.Println(id, left, right)
		return right.Cmp(id) >= 0 || id.Cmp(left) == 1
		// left --- id --- 0 or 0 --- id --- right
	}
}

// judge if Exclusive id belongs to (left, right)
func ExclusiveBetween(id *big.Int, left *big.Int, right *big.Int) bool {

	if right.Cmp(left) == 1 {
		//right > left
		// fmt.Println(left, "\n", right)
		return id.Cmp(left) == 1 && right.Cmp(id) == 1
	} else {
		return right.Cmp(id) == 1 || id.Cmp(left) == 1
		// left --- id --- 0 or 0 --- id --- right
	}

}
