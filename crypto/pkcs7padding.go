package pkcs7padding

func PKCS7Padding(data []byte) []byte {
	// 数据的长度
	dataLen := len(data)

	var bit16 int

	if dataLen%16 == 0 {
		bit16 = dataLen
	} else {
		// 计算补足的位数，填补16的位数，例如 10 = 16, 17 = 32, 33 = 48
		bit16 = int(dataLen/16+1) * 16
	}

	// 需要填充的数量
	paddingNum := bit16 - dataLen

	bitCode := byte(paddingNum)

	padding := make([]byte, paddingNum)
	for i := 0; i < paddingNum; i++ {
		padding[i] = bitCode

	}
	return append(data, padding...)
}

/**
 *	去除PKCS7的补码
 */
func UnPKCS7Padding(data []byte) []byte {
	dataLen := len(data)

	// 在使用PKCS7会以16的倍数减去数据的长度=补位的字节数作为填充的补码，所以现在获取最后一位字节数进行切割
	endIndex := int(data[dataLen-1])

	// 验证结尾字节数是否符合标准，PKCS7的补码字节只会是1-15的字节数
	if 16 > endIndex {

		// 判断结尾的补码是否相同 TODO 不相同也先不管了，暂时不知道怎么处理
		if 1 < endIndex {
			for i := dataLen - endIndex; i < dataLen; i++ {
				if data[dataLen-1] != data[i] {
					// fmt.Println("不同的字节码，尾部字节码:", data[dataLen-1], "  下标：", i, "  字节码：", data[i])
				}
			}
		}

		return data[:dataLen-endIndex]
	}

	// fmt.Println(endIndex)

	return nil
}
