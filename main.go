package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sync"
)

type Task struct {
	Content  string `json:"content"`
	Complete bool   `json:"complete"`
}

var (
	tasks = []Task{}
	mu    sync.Mutex
)

func listTasksHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if r.Method == http.MethodGet {
		// Отправляем список задач в формате JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
		return
	}

	// Если это не GET-запрос, возвращаем ошибку
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		tasks = append(tasks, task)
		mu.Unlock()

		w.WriteHeader(http.StatusCreated)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func completeTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var index int
		if err := json.NewDecoder(r.Body).Decode(&index); err != nil || index < 0 || index >= len(tasks) {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}

		mu.Lock()
		tasks[index].Complete = true
		mu.Unlock()

		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		var index int
		if err := json.NewDecoder(r.Body).Decode(&index); err != nil || index < 0 || index >= len(tasks) {
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}

		mu.Lock()
		tasks = append(tasks[:index], tasks[index+1:]...) // Удаляем задачу по индексу
		mu.Unlock()

		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func resetTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		mu.Lock()
		tasks = []Task{} // Сбрасываем все задачи
		mu.Unlock()

		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func renderTemplate(w http.ResponseWriter) {
	tmpl := template.Must(template.ParseFiles("template.html"))
	mu.Lock()
	defer mu.Unlock()
	tmpl.Execute(w, tasks)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w)
	})

	http.HandleFunc("/tasks", listTasksHandler)
	http.HandleFunc("/tasks/add", addTaskHandler)
	http.HandleFunc("/tasks/complete", completeTaskHandler)
	http.HandleFunc("/tasks/delete", deleteTaskHandler)
	http.HandleFunc("/tasks/reset", resetTasksHandler)

	http.ListenAndServe(":8080", nil)
}
