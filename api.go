package aqua

type Api struct{ Fixture }

type GetApi struct{ Api }
type PostApi struct{ Api }
type PutApi struct{ Api }
type PatchApi struct{ Api }
type DeleteApi struct{ Api }

// type bingoService struct {
// 	RestService

// 	bind aqua.DbHook `honor:CRUD`
// }

// func Bind() (conn, bolt.Bingo) {

// }
