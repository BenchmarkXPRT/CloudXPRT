/*******************************************************************************
* Copyright 2020 BenchmarkXPRT Development Community
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*******************************************************************************/

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spacemonkeygo/openssl"
)

const (
	ChunkSize = 10 * 1024
	TagSize   = 16
)

func EncryptFile(inFile, outFile *os.File, key, iv []byte) error {
	ctx, err := openssl.NewGCMEncryptionCipherCtx(len(key)*8, nil, key, iv)
	if err != nil {
		return fmt.Errorf("Failed making GCM encryption ctx: %v", err)
	}

	reader := bufio.NewReader(inFile)
	chunk := make([]byte, ChunkSize)
	for {
		chunkSize, err := reader.Read(chunk)
		if err == io.EOF || chunkSize == 0 {
			break
		} else if err != nil {
			return fmt.Errorf("Failed to read a chunk: %v", err)
		}

		encData, err := ctx.EncryptUpdate(chunk[:chunkSize])
		if err != nil {
			return fmt.Errorf("Failed to perform an encryption: %v", err)
		}

		if _, err := outFile.Write(encData); err != nil {
			return fmt.Errorf("Failed to write an encrypted data: %v", err)
		}
	}

	encData, err := ctx.EncryptFinal()
	if err != nil {
		return fmt.Errorf("Failed to finalize encryption: %v", err)
	}
	if _, err := outFile.Write(encData); err != nil {
		return fmt.Errorf("Failed to write a final encrypted data: %v", err)
	}

	tag, err := ctx.GetTag()
	if err != nil {
		return fmt.Errorf("Failed to get GCM tag: %v", err)
	}
	if _, err := outFile.Write(tag); err != nil {
		return fmt.Errorf("Failed to write a gcm tag: %v", err)
	}

	return nil
}

func EncryptString(input string, key, iv []byte) ([]byte, error) {
	ctx, err := openssl.NewGCMEncryptionCipherCtx(len(key)*8, nil, key, iv)
	if err != nil {
		return nil, fmt.Errorf("Failed making GCM encryption ctx: %v", err)
	}

	reader := strings.NewReader(input)
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	chunk := make([]byte, ChunkSize)
	for {
		chunkSize, err := reader.Read(chunk)
		if err == io.EOF || chunkSize == 0 {
			break
		} else if err != nil {
			return nil, fmt.Errorf("Failed to read a chunk: %v", err)
		}

		encData, err := ctx.EncryptUpdate(chunk[:chunkSize])
		if err != nil {
			return nil, fmt.Errorf("Failed to perform an encryption: %v", err)
		}

		if _, err := writer.Write(encData); err != nil {
			return nil, fmt.Errorf("Failed to write an encrypted data: %v", err)
		}
	}

	encData, err := ctx.EncryptFinal()
	if err != nil {
		return nil, fmt.Errorf("Failed to finalize encryption: %v", err)
	}
	if _, err := writer.Write(encData); err != nil {
		return nil, fmt.Errorf("Failed to write a final encrypted data: %v", err)
	}

	tag, err := ctx.GetTag()
	if err != nil {
		return nil, fmt.Errorf("Failed to get GCM tag: %v", err)
	}
	if _, err := writer.Write(tag); err != nil {
		return nil, fmt.Errorf("Failed to write a gcm tag: %v", err)
	}
	writer.Flush()

	return buf.Bytes(), nil
}

func DecryptFile(inFile, outFile *os.File, key, iv []byte, fileSize int) error {
	ctx, err := openssl.NewGCMDecryptionCipherCtx(len(key)*8, nil, key, iv)
	if err != nil {
		return fmt.Errorf("Failed making GCM decryption ctx: %v", err)
	}

	reader := bufio.NewReader(inFile)
	chunk := make([]byte, ChunkSize)
	tag := &bytes.Buffer{}
	totalChunkSize := 0

	for {
		chunkSize, err := reader.Read(chunk)
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("Failed to read an encrypted chunk: %v", err)
		}

		totalChunkSize += chunkSize

		if totalChunkSize > fileSize {
			d := totalChunkSize % fileSize

			d %= ChunkSize
			if d == 0 {
				d = chunkSize
			}

			tag.Write(chunk[chunkSize-d : chunkSize])
			chunkSize -= d
		}

		if chunkSize > 0 {
			data, err := ctx.DecryptUpdate(chunk[:chunkSize])
			if err != nil {
				return fmt.Errorf("Failed to perform a decryption: %v", err)
			}

			if _, err := outFile.Write(data); err != nil {
				return fmt.Errorf("Failed to write a decrypted data: %v", err)
			}
		}
	}

	if err := ctx.SetTag(tag.Bytes()); err != nil {
		return fmt.Errorf("Failed to set expected GCM tag: %v", err)
	}

	data, err := ctx.DecryptFinal()
	if err != nil {
		return fmt.Errorf("Failed to finalize decryption: %v", err)
	}
	if _, err := outFile.Write(data); err != nil {
		return fmt.Errorf("Failed to write a final decrypted data: %v", err)
	}

	return nil
}

func DecryptToString(input []byte, key, iv []byte) (string, error) {
	ctx, err := openssl.NewGCMDecryptionCipherCtx(len(key)*8, nil, key, iv)
	if err != nil {
		return "", fmt.Errorf("Failed making GCM decryption ctx: %v", err)
	}

	reader := bytes.NewReader(input)
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	chunk := make([]byte, ChunkSize)
	tag := &bytes.Buffer{}
	totalChunkSize := 0
	inputSize := len(input) - TagSize

	for {
		chunkSize, err := reader.Read(chunk)
		if err == io.EOF {
			break
		} else if err != nil {
			return "", fmt.Errorf("Failed to read an encrypted chunk: %v", err)
		}

		totalChunkSize += chunkSize

		if totalChunkSize > inputSize {
			/*
				d := totalChunkSize % inputSize
				if d < TagSize {
					d = TagSize
				}
				d %= ChunkSize
				if d == 0 {
					d = chunkSize
				}
			*/
			d := TagSize
			tag.Write(chunk[chunkSize-d : chunkSize])
			chunkSize -= d
		}

		if chunkSize > 0 {
			data, err := ctx.DecryptUpdate(chunk[:chunkSize])
			if err != nil {
				return "", fmt.Errorf("Failed to perform a decryption: %v", err)
			}

			if _, err := writer.Write(data); err != nil {
				return "", fmt.Errorf("Failed to write a decrypted data: %v", err)
			}
		}
	}

	if err := ctx.SetTag(tag.Bytes()); err != nil {
		return "", fmt.Errorf("Failed to set expected GCM tag: %v", err)
	}

	data, err := ctx.DecryptFinal()
	if err != nil {
		return "", fmt.Errorf("Failed to finalize decryption: %v", err)
	}
	if _, err := writer.Write(data); err != nil {
		return "", fmt.Errorf("Failed to write a final decrypted data: %v", err)
	}
	writer.Flush()

	return buf.String(), nil
}
