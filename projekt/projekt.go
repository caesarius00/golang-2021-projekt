package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const key = "password"

func clearConsole() {
	cmd := exec.Command(("clear"))
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func remove(slice []string, number int) []string {
	return append(slice[:number], slice[number+1:]...)
}

func askQuestion(yourAnswer []string, set []string, right []string, answers []string, qNumber int, iterator int) int {
	fmt.Println(set[qNumber])
	fmt.Println(answers[qNumber])
	var givenAnswer string
	rightAnswer := right[qNumber]
	regexSmall, _ := regexp.Compile("[a-d]")
	regexBig, _ := regexp.Compile("[A-D]")
	for err := 0; !(regexSmall.MatchString(givenAnswer) || regexBig.MatchString(givenAnswer)) || len(givenAnswer) != 1; err++ {
		if err > 0 {
			fmt.Println("Nie ma takiej odpowiedzi, spróbuj jeszcze raz!")
		}
		fmt.Scanln(&givenAnswer)
	}

	if regexSmall.MatchString(givenAnswer) {
		givenAnswer = strings.ToUpper(givenAnswer)
	}
	yourAnswer[iterator] = givenAnswer
	remove(set, qNumber)
	remove(right, qNumber)
	remove(answers, qNumber)
	if givenAnswer == rightAnswer {
		return 1
	}
	return 0
}

func showQuestion(yourAnswer []string, set []string, right []string, answers []string, qNumber int, iterator int) {
	fmt.Println()
	fmt.Println(set[qNumber])
	fmt.Println(answers[qNumber])
	fmt.Println("Twoja odpowiedź: ", yourAnswer[iterator])
	fmt.Println("Poprawna odpowiedź: ", right[qNumber])

	remove(set, qNumber)
	remove(right, qNumber)
	remove(answers, qNumber)
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, _ := cipher.NewGCM((block))
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.New(rand.NewSource(time.Now().UnixNano())), nonce)
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, _ := gcm.Open(nil, nonce, ciphertext, nil)
	return plaintext
}

func encryptFile(filename string, passphrase string, dest string) {
	in, _ := ioutil.ReadFile(filename)
	fmt.Println(string(in))
	debug := strings.Index(filename, "/")
	if debug > 0 {
		f, _ := os.Create(dest + filename[debug:])
		defer f.Close()
		f.Write(encrypt([]byte(in), passphrase))
	} else {
		f, _ := os.Create(dest + filename)
		defer f.Close()
		f.Write(encrypt([]byte(in), passphrase))
	}
	os.Remove(filename)
}

func decryptFile(filename string, passphrase string, from string) []byte {
	data, _ := ioutil.ReadFile(from + filename)
	return decrypt(data, passphrase)
}

func addTest(name string) {
	if _, err := os.Stat("zaszyfrowane"); os.IsNotExist(err) {
		os.Mkdir("zaszyfrowane", 0777)
	}
	encryptFile("nowe/"+name, key, "zaszyfrowane/")
	fmt.Println("Dodano test", name[:len(name)-4])
}

func getTest(name string, questionSet []string, answerSet []string, rightAnswerSet []string) ([]string, []string, []string) {
	content := string(decryptFile(name, key, "zaszyfrowane/"))
	//counter przechowuje zapisaną ilość linii = ~ ilość pytań * 3
	//fmt.Println(content)
	counter := 0
	beginning := 1
	newLine := 0
	answerGet := 0
	var answers string
	for i, sign := range content {
		if sign == 10 {
			counter++
			newLine = 1
			answerGet = 0
		} else {
			switch counter % 3 {
			case 0:
				if newLine == 1 {
					questionSet = append(questionSet, string(sign))
					answerSet = append(answerSet, answers+"\n")
					answers = ""
					newLine = 0
				} else if beginning == 1 {
					questionSet = append(questionSet, string(sign))

					beginning = 0
				} else if content[i+1] != 10 {
					questionSet[counter/3] += string(sign)
				}
			case 1:
				if answerGet == 0 {
					rightAnswerSet = append(rightAnswerSet, string(content[i-2]))
					answerGet = 1
				}
				if sign == '.' && (content[i-1] == 'B' || content[i-1] == 'C' || content[i-1] == 'D') {
					answers = answers[:len(answers)-1] + "\n" + answers[len(answers)-1:]

				}
				answers += string(sign)
			}
		}
	}
	answerSet = append(answerSet, answers)
	return questionSet, answerSet, rightAnswerSet
}

func findFiles(dir string) []string {
	filesAll, _ := ioutil.ReadDir(dir)
	files := make([]string, 0)
	for _, f := range filesAll {
		//interesują mnie tylko pliki formatu.txt,
		//więc tylko takie pliki przyjmuję jako te do odczytania
		temp := f.Name()
		extension := temp[len(temp)-4:]
		if extension == ".txt" {
			files = append(files, f.Name())
		}
	}
	return files
}

func fileChoice(dir string) (int, []string) {
	files := findFiles(dir)
	choice := -1
	fmt.Println("Wybierz spośród dostępnych:")
	for i := 0; i < len(files); i++ {
		fmt.Println(i+1, "-", files[i][:len(files[i])-4])
	}
	if len(files) == 0 {
		fmt.Println("BRAK DANYCH")
	} else {
		for err := 0; choice > len(files) || choice < 0; err++ {
			if err > 0 {
				fmt.Println("Zła opcja, wybierz ponownie:")
			}
			fmt.Scanln(&choice)
		}
	}
	return choice, files
}

func saveData(user string, rightCounter int, max int, quizName string) {
	dt := time.Now()
	files := findFiles("wyniki")
	gameName := user + "-" + quizName
	flag := 0
	for i := 0; i < len(files); i++ {
		if gameName == files[i] {
			flag++
		}
	}
	if flag == 1 {
		saved := string(decryptFile(gameName, key, "wyniki/"))
		saved += "\n" + dt.Format("01-02-2006 15:04:05 Monday") + " WYNIK: " + strconv.Itoa(rightCounter) + "/" + strconv.Itoa(max)
		file, _ := os.Create(gameName)
		defer file.Close()
		file.WriteString(saved)
		encryptFile(gameName, key, "wyniki/")
		os.Remove(gameName)
	} else {
		file, _ := os.Create(gameName)
		defer file.Close()
		add := "WYNIKI GRACZA " + user + " W " + quizName[:len(quizName)-4] + "\n\n" + dt.Format("01-02-2006 15:04:05 Monday") + " WYNIK: " + strconv.Itoa(rightCounter) + "/" + strconv.Itoa(max)
		file.WriteString(add)
		encryptFile(gameName, key, "wyniki/")
		os.Remove(gameName)
	}
}

//Stworzyłem przykładowe testy do sprawdzenia programu

/*W celu dodania własnego pliku z quizem należy przestrzegać następujących wskazań:
Pytania powinny się znajdować od pierwszej lini w formacie:
	1. linia - pytanie i na koniec znak odpowiedzi (A, B, C lub D)
	2. linia - odpowiedzi w formacie A. xyz B. abc ... D. pty
	3. linia pusta lub nieistotna
*/

func main() {
	fmt.Println("Zaraz rozpocznie się Twój test")
	choice := -1
	var user string
	for choice != 0 {
		fmt.Println("M E N U")
		fmt.Println("0 - WYJŚCIE")
		fmt.Println("1 - dodaj test (szyfrowanie pliku")
		fmt.Println("2 - rozpocznij test")
		fmt.Println("3 - zobacz wyniki")
		fmt.Scanln(&choice)
		switch choice {
		case 0:
			break
		case 1:
			if _, err := os.Stat("nowe"); os.IsNotExist(err) {
				os.Mkdir("nowe", 0777)
			}
			files1 := make([]string, 0)
			choice, files1 = fileChoice("nowe")
			addTest(files1[choice-1])
			choice = -1
		case 2:
			if _, err := os.Stat("zaszyfrowane"); os.IsNotExist(err) {
				os.Mkdir("zaszyfrowane", 0777)
			}
			fmt.Println("Podaj nazwę gracza: ")
			fmt.Scanln(&user)
			files2 := make([]string, 0)
			choice, files2 = fileChoice("zaszyfrowane")
			if choice != 0 {
				var rightCounter int
				questionSet := make([]string, 0)
				answerSet := make([]string, 0)
				rightAnswerSet := make([]string, 0)

				questionSet, answerSet, rightAnswerSet = getTest(files2[choice-1], questionSet, answerSet, rightAnswerSet)

				//tworzę kopie slice'ów, aby nie stracić oryginalnie wczytanych danych
				setCopy := make([]string, len(questionSet))
				copy(setCopy, questionSet)
				rightCopy := make([]string, len(questionSet))
				copy(rightCopy, rightAnswerSet)
				answerCopy := make([]string, len(questionSet))
				copy(answerCopy, answerSet)
				yourAnswer := make([]string, len(questionSet))

				//chcę przechować kolejność zadawania pytań (zadawane są losowo)
				//aby wyświetlić na koniec w tej samej kolejności wraz z poprawnymi odpowiedziami
				questionOrder := make([]int, len(questionSet))

				clearConsole()

				for i := 0; i < len(questionSet); i++ {
					if i+1 != len(questionSet) {
						r := rand.New(rand.NewSource(time.Now().UnixNano()))
						order := r.Int() % (len(questionSet) - i - 1)
						questionOrder[i] = order
						rightCounter += askQuestion(yourAnswer, setCopy, rightCopy, answerCopy, order, i)
					} else {
						questionOrder[i] = 0
						rightCounter += askQuestion(yourAnswer, setCopy, rightCopy, answerCopy, 0, i)
					}
					clearConsole()
				}
				if rightCounter == len(questionSet) {
					fmt.Println("GRATULACJE! Wszystko dobrze")
				} else {
					fmt.Println("ODPOWIEDZI: ")
					for i := 0; i < len(questionSet); i++ {
						showQuestion(yourAnswer, questionSet, rightAnswerSet, answerSet, questionOrder[i], i)
					}
					fmt.Println("Udzielono", rightCounter, "/", len(questionSet), "poprawnych odpowiedzi")
					fmt.Println("^^^ Powyżej pytania wraz z odpowiedziami ^^^ ")
					saveData(user, rightCounter, len(questionSet), files2[choice-1])
					fmt.Println("Dane o rozgrywce zostały zaktualizowane")
				}
				fmt.Println("Spróbuj inny quiz! ")

				//resetuję choice na -1, aby wszedł do poniższej pętli,
				//gdy skończy całą zewnętrzną pętlę jeden (wiele) raz(y)
			}
			choice = -1
		case 3:
			if _, err := os.Stat("wyniki"); os.IsNotExist(err) {
				os.Mkdir("wyniki", 0777)
			}
			files3 := make([]string, 0)
			choice, files3 = fileChoice("wyniki")
			if choice > 0 {
				fmt.Println(string(decryptFile(files3[choice-1], key, "wyniki/")))
			}
			choice = -1
		}
	}
	fmt.Println("Do zobaczenia! ;-)")
}
