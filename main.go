package main

import (
	"context"
	"strconv"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
	"log"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/joho/godotenv"
	"github.com/gookit/color"
)



type MongoDatabaseAndCollections struct {
	db *mongo.Database
	taskCollection *mongo.Collection	
}

var mongoDB *MongoDatabaseAndCollections
var once sync.Once	


type Task struct {
	Name string `bson:"name" json:"name"`
	Description string `bson:"Description" json:"Description"`
	Completed bool `bson:"completed" json:"completed"`
	Duration time.Duration `bson:"duration" json:"duration"`
	Deadline time.Time `bson:"deadline" json:"deadline"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}


func (t *Task) create()   error {
	t.Completed = false 
	t.CreatedAt = time.Now() 
	t.UpdatedAt = time.Now() 

	_,err := mongoDB.taskCollection.InsertOne(context.TODO(),t)
	return err

}


func  updateTaskByName( name string , updatePayload bson.D)  (  error ){
	filter := bson.D{{"name", name}}		
	var err error 
	if _, err = mongoDB.taskCollection.UpdateOne(context.TODO(),filter,updatePayload) ; err != nil {
		return err 
	}
	return  nil 
}


func  completeTask( name string)  (  error ){
	updatePayload :=  bson.D{{"$set", bson.D{{"completed", true}}}}	
	return updateTaskByName(name , updatePayload)
}


func  deleteTask( name string)  (  error ){
	filter := bson.D{{"name", name}}		
	var err error 
	if _, err = mongoDB.taskCollection.DeleteOne(context.TODO(),filter) ; err != nil {
		return err 
	}
	return  nil 
}

func getTasks(  filter bson.D) ( []*Task,  error ) {
	
	var tasks []*Task
	curr := &mongo.Cursor{}
	var err error
	if curr , err = mongoDB.taskCollection.Find(context.TODO(),filter); err != nil {
		return nil, err
	}

	for curr.Next(context.TODO()){
		var task Task 

		if err = curr.Decode(&task) ; err != nil {
			return nil , err
		}

		tasks = append(tasks, &task)
	}

	curr.Close(context.TODO())

	if len(tasks) == 0 {
		return tasks, mongo.ErrNoDocuments
	}
	
	return tasks, nil 
}


func  getAllTasks()  ( []*Task,  error ){

	filter := bson.D{{}}
	return getTasks(filter) 
	
}

func  getPendingTasks()  ( []*Task,  error ){

	filter := bson.D{{"completed",false}}
	return getTasks(filter) 
	
}
func  getCompletedTaks()  ( []*Task,  error ){

	filter := bson.D{{"completed",true}}
	return getTasks(filter) 
	
}




func (t *Task) complete()   error {
	t.Completed = false 
	t.CreatedAt = time.Now() 
	t.UpdatedAt = time.Now() 

	_,err := mongoDB.taskCollection.InsertOne(context.TODO(),t)
	return err
}


func printTasks( tasks[]*Task  ) {
	for i , t := range tasks {
		if t.Completed {
			color.Green.Printf(" %d : %s\n",i+1,t.Name)
		}else{
			color.Yellow.Printf(" %d : %s\n",i+1,t.Name)
		}
	}
}

func integersToDuration(h int , m int , s int) time.Duration {
	return time.Hour * time.Duration(h) + time.Minute * time.Duration(m) + time.Second * time.Duration(s)  		
}

func stringToIntOrZero( s string ) ( int,error) {
	if s == "" {
		return 0 , nil
	} 
	return strconv.Atoi(s)
}

var AddTaskCommand =  &cli.Command{
	Name: "add",
	Aliases: []string{"a"},
	Usage: "add a new task",
	Action : func(c *cli.Context) error{
		name := c.Args().Get(0)
		description :=  c.Args().Get(1)
		var durationH int
		var durationM int
		var durationS int
		var err error 
		var duration time.Duration		

		if durationH , err = stringToIntOrZero(c.Args().Get(2)) ; err != nil {
	return err
		}

		if durationM , err = stringToIntOrZero(c.Args().Get(3)) ; err != nil {
	return err
		}
	
		if durationS , err = stringToIntOrZero(c.Args().Get(4)) ; err != nil {
	return err
		}
	

		if name == "" {
			return errors.New("the task name is empty .")
		}
	
		if durationH == 0 && durationM == 0 && durationS == 0 {
			duration = time.Hour * time.Duration(1)
		}

		duration = integersToDuration(durationH,durationM,durationS)

		task := &Task{
			Name : name ,
			Description : description,
			Deadline : time.Now().Add(duration),
			Duration : duration , 
		}
		return task.create()		
	},
}

var GetAllCommand =  &cli.Command{
	Name: "getall",
	Aliases: []string{"ga"},
	Usage: "get all the tasks",
	Action : func(c *cli.Context) error{
		tasks ,err := getAllTasks()	
		if err != nil {
			if err == mongo.ErrNoDocuments {
				fmt.Println("There is no task")	
				return nil
			}
			return err	
		}		
		printTasks(tasks)
		return nil	
	},
}


var GetPendingTasksCommand =  &cli.Command{
	Name: "getpend",
	Aliases: []string{"gp"},
	Usage: "get all the pending tasks",
	Action : func(c *cli.Context) error{
		tasks ,err := getPendingTasks()	
		if err != nil {
			if err == mongo.ErrNoDocuments {
				fmt.Println("There is no task")	
				return nil
			}
			return err	
		}		
		printTasks(tasks)
		return nil	
	},
}

var GetCompletedTasksCommand =  &cli.Command{
	Name: "getcomp",
	Aliases: []string{"gc"},
	Usage: "get all the completed tasks",
	Action : func(c *cli.Context) error{
		tasks ,err := getCompletedTaks()	
		if err != nil {
			if err == mongo.ErrNoDocuments {
				fmt.Println("There is no task")	
				return nil
			}
			return err	
		}		
		printTasks(tasks)
		return nil	
	},
}


var CompleteTaskCommand =  &cli.Command{
	Name: "complete",
	Aliases: []string{"c"},
	Usage: "mark a task as completed",
	Action : func(c *cli.Context) error{
		name := c.Args().Get(0)

		if name == "" {
			return errors.New("the task name is empty .")
		}
		
		return completeTask(name)	
	},
}



var DeleteTaskCommand =  &cli.Command{
	Name: "del",
	Aliases: []string{"d"},
	Usage: "delete a task",
	Action : func(c *cli.Context) error{
		name := c.Args().Get(0)

		if name == "" {
			return errors.New("the task name is empty .")
		}
		
		return deleteTask(name)		
	},
}


func loadEnvVariables()error{
	if err := godotenv.Load() ; err != nil {
		return err
	}	
	return nil 
}


func initDB()  error{
	var err error = nil 					
	var mongoClient *mongo.Client 
	var currErr error					
	once.Do(func(){
		uri := os.Getenv("MONGODB_CONTAINER_URI")
		if uri == "" {
			err = errors.New("you must set your MONGODB_URI in the .env file")
			

		}else if mongoClient ,currErr = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri)); currErr != nil {
			err = currErr	

		}else if currErr = mongoClient.Ping( context.TODO(), nil); currErr != nil {
			err = currErr	
		}else{
			db := mongoClient.Database("task-db")

			//fmt.Printf("creating the database %s\n",db.Name())
		
			taskCollection := db.Collection("Tasks")
			mongoDB  = &MongoDatabaseAndCollections{
				db,
				taskCollection, 
			}	

		}
				

	})
	//fmt.Printf(" initDB error : %v\n",err)
	return err
}


func disconnectDb() error {
	if err := mongoDB.db.Client().Disconnect(context.TODO()) ; err != nil {
		return err
	}
	return nil 

}




func main(){

	if err := loadEnvVariables() ; err != nil {
		log.Fatal(err)
	}

	if err := initDB() ; err != nil {
		log.Fatal(err)
	}
	
	defer disconnectDb()

	app := &cli.App{
		Name : "ToDoList",
		Description : "a simple Todo List CLI tool ",
		Commands : []*cli.Command{
			AddTaskCommand,
			GetAllCommand,
			CompleteTaskCommand,	
			GetPendingTasksCommand,
			GetCompletedTasksCommand,
			DeleteTaskCommand,
		},
		/*Action : func(c *cli.Context) error {
			fmt.Println("Bonjour")
			
			return nil
		},*/	

	}	
	

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
