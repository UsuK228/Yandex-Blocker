package main

import (
	"YandexBlocker/icon"
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	st "github.com/getlantern/systray"
	pc "github.com/shirou/gopsutil/v3/process"
	rg "golang.org/x/sys/windows/registry"
)

var BlockerEnabler = true // Коварка для переключения состояния программы
var yNames = []string{    // Список с возможными процессами
	"Yandex.exe",
	"browser.exe",
	"YandexSetup.exe",
	"YandexPackSetup.exe",
}
var yLinks = []string{ // Список возможных ссылок
	"yandex.ru",
	"www.yandex.ru",
	"browser.yandex.ru",

	"yandex.net",
	"www.yandex.net",
	"cdn.yandex.net",
	"downloader.yandex.net",
	"download.cdn.yandex.net",

	"ya.ru",
	"www.ya.ru",
}
var yPaths = []string{
	filepath.Join(os.Getenv("APPDATA"), "Yandex"),
	filepath.Join(os.Getenv("LOCALAPPDATA"), "Yandex"),
	filepath.Clean("C:\\Program Files (x86)\\Yandex"),
}

const hostsPath = "C:/Windows/System32/drivers/etc/hosts" // Ну файл hosts что еще говорить
const marker = "# BY YANDEX BLOCKER"                      // Нужен, чтобы распознавать строки в файле hosts, вот такие костыли

func main() {
	RunAsAdmin() // Запускает прогу от имени администратора
	log.SetPrefix("[DEBUG] ")
	log.Println("Стартуем...")
	st.Run(Worker, Exit)
}

func Worker() {
	// Инициализация воркера
	log.Println("Инициализируем воркера...")
	// Ставим иконку, титл и подсказКО
	st.SetTitle("Блокиратор Яндекс")
	st.SetTooltip("Блокиратор Яндекс")
	st.SetIcon(icon.Mark)

	// Добавляем кнопки, ксожалению иконки поставить не получилось, обыдло
	mQuit := st.AddMenuItem("Выход", "Завершает программу")
	mEnabler := st.AddMenuItem("Вкл/Выкл", "Переключить режим программы, галочка означает что блокиратор работает")

	// Логика переключения состояния программы
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				st.Quit()
				return
			case <-mEnabler.ClickedCh:
				if !BlockerEnabler {
					BlockerEnabler = true
					st.SetIcon(icon.Mark)
				} else {
					BlockerEnabler = false
					st.SetIcon(icon.Cross)
				}
				log.Println("Переключено состояние программы на", BlockerEnabler)
			}
		}
	}()

	// Логика программы
	go func() {
		for {
			BlockDownload()
			time.Sleep(3 * time.Second)
			KillProcesses()
			time.Sleep(3 * time.Second)
			DeleteFiles()
			time.Sleep(3 * time.Second)
		}
	}()

	log.Println("Воркер инициализирован!")
}

func RunAsAdmin() {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		log.Println("Программа не запущена от имени администратора, перезапускаем...")
		exe, err := os.Executable()
		if err != nil {
			log.Fatalln("Ошибка получения файла блокиратора (КАК?)")
		}
		cmd := exec.Command("powershell", "-Command", "Start-Process", "-FilePath", exe, "-Verb", "RunAs")
		_ = cmd.Run()

		os.Exit(0)
	}
}

func BlockDownload() {
	// Блокиратор загрузки файлов и доступа на сайт АКА редачер файла hosts
	// Читаем hosts
	file, err := os.Open(hostsPath)
	if err != nil {
		log.Println("Ошибка открытия файла hosts")
	}

	scanner := bufio.NewScanner(file)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Запись в hosts
	var newLines []string
	for _, line := range lines {
		if !strings.Contains(strings.TrimSpace(line), marker) {
			newLines = append(newLines, line)
		}
	}

	if BlockerEnabler {
		// newLines = append(newLines, "") // Пустая строка
		for _, d := range yLinks {
			newLines = append(newLines, fmt.Sprintf("127.0.0.1 %s %s", d, marker))
		}
	}

	out, err := os.OpenFile(hostsPath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("Ошибка записи в файл hosts")
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	for _, line := range newLines {
		if _, err := out.WriteString(line + "\n"); err != nil {
			log.Println("Ошибка записи в файл hosts")
		}
	}
}

func DeleteFiles() {
	// Удаление файлов Яндекса
	if BlockerEnabler {
		// Удаляем из реестра
		err := rg.DeleteKey(rg.LOCAL_MACHINE, "SYSTEM\\ControlSet001\\Services\\YandexBrowserService")
		if err != nil {
			log.Println("Ошибка удаления из реестра (Возможно нет ключа)")
		}

		// Удаляем папки Яндекса
		for _, path := range yPaths {
			err := os.RemoveAll(path)
			if err != nil {
				log.Println("Ошибка удаления папки Яндекса (Возможно нет папки)", path)
			}
		}
	}
}

func KillProcesses() {
	// Убиваем процесс(ы) Яндекса
	if BlockerEnabler {
		processes, err := pc.Processes()
		if err != nil {
			log.Fatalln("Ошибка:", err)
		}

		for _, p := range processes {
			pName, err := p.Name()
			if err != nil {
				continue // Мог завершиться
			}

			for _, tName := range yNames {
				if strings.EqualFold(pName, tName) {
					log.Println("Убиваем процесс:", pName, p.Pid)
					err := p.Kill()
					if err != nil {
						log.Println("Не удалось убить процесс: ", pName, p.Pid)
					}
				}
			}
		}
	}

}

func Exit() {
	log.Println("Выходим...")
	os.Exit(1)
}
