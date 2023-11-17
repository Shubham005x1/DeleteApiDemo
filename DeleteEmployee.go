package content

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/logging"
	"github.com/gorilla/mux"
)

type Employee struct {
	ID        string `firestore:"id" json:"id"`
	FirstName string `firestore:"firstname" json:"firstname"`
	LastName  string `firestore:"lastname" json:"lastname"`
	Email     string `firestore:"email" json:"email"`
	Password  string `firestore:"password" json:"-"`
	PhoneNo   string `firestore:"phoneNo" json:"phoneNo"`
	Role      string `firestore:"role" json:"role"`
}

var (
	client     *firestore.Client
	logClient  *logging.Client
	onceClient sync.Once
)

func InitializeFirestore() {
	onceClient.Do(func() {
		ctx := context.Background()

		// Initialize Firestore with the service account key
		var err error
		client, err = firestore.NewClient(ctx, "takeoff-task-3")
		if err != nil {
			log.Fatalf("Failed to create Firestore client::: %v", err)
		}
	})
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/employees", DeleteEmployee).Methods("Delete")
	log.Println("Server started on :8082")
	log.Fatal(http.ListenAndServe(":8082", router))
}

func DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	// Get employee ID from the query parameter
	ctx := context.Background()
	id := r.URL.Query().Get("id")
	logClient, _ = logging.NewClient(ctx, "takeoff-task-3")
	defer logClient.Close()
	logger := logClient.Logger("my-log")

	// Log an entry indicating that the CreateEmployee function has started.
	logger.Log(logging.Entry{
		Payload:  "Delete function started",
		Severity: logging.Info,
	})
	// Check if the id parameter is empty
	if id == "" {
		http.Error(w, "Missing 'id' query parameter", http.StatusBadRequest)
		return
	}

	// Initialize Firestore client
	InitializeFirestore()
	collectionRef := client.Collection("employees")
	query := collectionRef.Where("id", "==", id)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve employee document: %v", err), http.StatusInternalServerError)
		return
	}

	if len(docs) == 0 {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	docid := docs[0].Ref.ID

	// Delete the employee from Firestore
	docRef := collectionRef.Doc(docid)

	_, err = docRef.Delete(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete employee: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Employee deleted successfully"))
	logger.Log(logging.Entry{
		Payload:  "Employee deleted successfully!",
		Severity: logging.Info,
	})

}
