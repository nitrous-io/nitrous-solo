package escapecsi

// Refer to https://en.wikipedia.org/wiki/ANSI_escape_code
func EscapeCSI(data []byte) []byte {
	result := make([]byte, len(data))
	size := 0

	for i := 0; i < len(data); i++ {
		if i+5 >= len(data) {
			result[size] = data[i]
			size++
			continue
		}

		if data[i] == 27 {
			if data[i+1] == 91 {
				var j = i + 2
				if data[j] >= '0' && data[j] <= '9' {
					for data[j] >= '0' && data[j] <= '9' {
						j++
						if j >= len(data) {
							break
						}
					}

					if j+1 >= len(data) {
						result[size] = data[i]
						size++
						continue
					}

					if data[j] == ';' {
						if data[j+1] >= '0' && data[j+1] <= '9' {
							j++
							for data[j] >= '0' && data[j] <= '9' {
								j++
								if j >= len(data) {
									break
								}
							}

							if j >= len(data) {
								result[size] = data[i]
								size++
								continue
							}
						} else {
							result[size] = data[i]
							size++
							continue
						}
					}

					if data[j] != 'm' && (data[j] >= 'a' && data[j] <= 'z') || (data[j] >= 'A' && data[j] <= 'Z') {
						i = j
						continue
					}
				}
			}
		}

		result[size] = data[i]
		size++
	}

	newData := make([]byte, size)
	copy(newData, result[:size])

	return newData
}
