package basex

import (
	"bytes"
	"fmt"
	"math/big"
)

// Encoder は符号化/復号化するためのタイプになります
type Encoder struct {
	chars []byte
	demap map[byte]int64
}

// NewEncoder は指定した符号列にEncodeするためのEncoderを生成します
func NewEncoder(chars string) (*Encoder, error) {
	if len(chars) < 2 {
		return nil, fmt.Errorf("符号文字は1文字以上指定してください")
	}
	encoder := &Encoder{}

	// 符号文字をコピーする
	encoder.chars = make([]byte, len(chars))
	copy(encoder.chars[:], chars[:])

	// 符号文字をバイト値とindexを覚えておく
	encoder.demap = map[byte]int64{}
	for i, b := range []byte(chars) {
		encoder.demap[b] = int64(i)
	}

	return encoder, nil
}

var zero = big.NewInt(0)

// Encode は符号化する
func (encoder *Encoder) Encode(src []byte) string {
	// 結果のバッファを格納
	res := &bytes.Buffer{}

	// 先頭の0x00バイトを先に埋めておく
	// 理由は数値に変換する際に欠落する先頭の0byteの符号化
	for _, c := range src {
		if c != 0x00 {
			break
		}
		// bytes.WriteByteはエラーは返すが見かけだけなのでハンドリングはスルーする
		res.WriteByte(encoder.chars[0])
	}

	// バイト列を数値に変換して基数を元にmodしたindexを元に符号化していく
	num := new(big.Int).SetBytes(src)
	div := big.NewInt(int64(encoder.Base()))
	mod := &big.Int{}
	buf := &bytes.Buffer{}

	for num.Cmp(zero) != 0 {
		num.DivMod(num, div, mod)
		buf.WriteByte(encoder.chars[mod.Int64()])
	}

	// 最後に符合結果が逆になっているので並び替えて結果に出力
	bufBytes := buf.Bytes()
	for i := len(bufBytes) - 1; 0 <= i; i-- {
		res.WriteByte(bufBytes[i])
	}
	return res.String()
}

// Decode は符号化された文字列を元のバイト配列に復号化する
func (encoder *Encoder) Decode(encoded string) ([]byte, error) {
	src := []byte(encoded)
	buf := new(bytes.Buffer)

	// 先頭の0byte文字をバイナリに変換
	startIdx := 0
	for i, c := range src {
		if c != encoder.chars[0] {
			startIdx = i
			break
		}
		buf.WriteByte(0x00)
	}

	num := big.NewInt(0)
	div := big.NewInt(int64(encoder.Base()))

	for _, c := range src[startIdx:] {
		charIdx, ok := encoder.demap[c]
		if !ok {
			return nil, fmt.Errorf("復号化できない文字がふくまれています: %v", c)
		}
		num.Add(num.Mul(num, div), big.NewInt(charIdx))
	}
	buf.Write(num.Bytes())

	return buf.Bytes(), nil
}

// Base はEncoderの基数を取得する
func (encoder *Encoder) Base() int {
	return len(encoder.chars)
}
