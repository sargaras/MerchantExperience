package processing

import (
	"database/sql"
	"fmt"

	"github.com/tealeg/xlsx"
)

//краткая статистика
// added - количество созданных товаров
// updated - количество обновленных товаров
// deleted - количество удаленных товаров
// wrong - количество неправильных строк
type declaration struct {
	added   uint
	updated uint
	deleted uint
	wrong   uint
}

// структура, возвращающая результат с проверкой на корректность данных
type xlsxData struct {
	product Product
	correct bool
}

//проверка на корректность считанных данных
func IsCorrect(loffer_id, lname, lprice, lquantity, lavailable string) xlsxData {

	isCorrect := true
	prod := Product{}
	//проверка на положительность чисел
	prod.offer_id, isCorrect = IsCorrectOfferId(loffer_id)

	//проверка на отсутсвите в начале чисел (товар не должен начинаться с чисел?)
	prod.name, isCorrect = IsCorrectName(lname)

	//проверка на отсутствие знаков и букв лишних в числе с плаваюещей точкой
	prod.price, isCorrect = IsCorrectPrice(lprice)

	//проверка на положительность чисел
	prod.quantity, isCorrect = IsCorrectQuantity(lquantity)

	//проверка на правильной записи типа bool
	prod.available, isCorrect = IsCorrectAvailable(lavailable)

	return xlsxData{prod, isCorrect}
}

//почему-то Баг с отсутсвием row до сих пор не пофиксили -_-
func ReadDataFromXLSX(exelFileName string) []xlsxData {
	xlsxDatas := []xlsxData{}

	xlFile, err := xlsx.OpenFile(exelFileName)
	checkErr(err)
	for _, sheet := range xlFile.Sheets {
		fmt.Println(&sheet)
		for row := 0; row != sheet.MaxRow; row++ {

			//На тот случай, если у нас данные сдвинуты в таблице по строке
			position := 0
			loffer_id, _ := sheet.Cell(row, position+0)
			for loffer_id == nil {
				position++
			}

			// loffer_id, _ := sheet.Cell(row, position+0)
			lname, _ := sheet.Cell(row, position+1)
			lprice, _ := sheet.Cell(row, position+2)
			lquantity, _ := sheet.Cell(row, position+3)
			lavailable, _ := sheet.Cell(row, position+4)

			xlsxData := IsCorrect(loffer_id.Value, lname.Value, lprice.Value, lquantity.Value, lavailable.Value)
			xlsxDatas = append(xlsxDatas, xlsxData)
		}
	}
	return xlsxDatas
}

/*
Общее обновление согласно поданным данным;
   declaration - статистика по данным:
added - добавлено
updated - обновлено
deleted - удалено
wrong - ошиблись
*/
func DelegateRequest(db *sql.DB, seller_id uint64, Products []xlsxData) declaration {
	addForProducts := []Product{}
	updateForProducts := []Product{}
	deleteForProducts := []Product{}
	//rensponsibilities := getViewRensposibility(db)
	var wrong uint = 0

	for _, value := range Products {
		isUpdated := false //проверка на обновление данных

		//проверка на корректность данных
		if value.correct == false {
			wrong++
			continue
		}

		rensponsibilities := LocalSelect(db, seller_id, value.product.offer_id, value.product.name)

		//обновление данных происходит в том случае, если указанный id продавца совпадает с id продавца из БД
		for _, Rensponsibility := range rensponsibilities {
			if Rensponsibility.seller.offer_id == value.product.offer_id && seller_id == Rensponsibility.seller.seller_id {
				//обновление данных
				updatedProduct := Product{}
				if value.product.available == true {
					//просто занимаемся сложением данных
					updatedProduct.quantity = Rensponsibility.product.quantity + value.product.quantity
					updatedProduct.offer_id = value.product.offer_id
					updatedProduct.name = value.product.name
					updatedProduct.price = value.product.price
					updatedProduct.available = true
					updateForProducts = append(updateForProducts, updatedProduct)
				}

				if value.product.available == false {
					updatedProduct.quantity = Rensponsibility.product.quantity - value.product.quantity
					if updatedProduct.quantity > 0 {
						//если товаров больше 0, то просто обновляем данные
						updatedProduct.offer_id = value.product.offer_id
						updatedProduct.name = value.product.name
						updatedProduct.price = value.product.price
						updatedProduct.available = value.product.available
						updateForProducts = append(updateForProducts, updatedProduct)
					} else if updatedProduct.quantity == 0 {
						// если товаров не осталось, то нужно просто удалить из БД
						deleteForProducts = append(deleteForProducts, Rensponsibility.product)
					} else if updatedProduct.quantity < 0 {
						//если в excel указано, что у нас больше товаров идет на удаление, то тут какая-то ошибка
						wrong++
					}
				}
				isUpdated = true
				//break
			}
		}

		if isUpdated == false {
			if value.product.available == true {
				addForProducts = append(addForProducts, value.product)
			} else {
				wrong++
			}
		}
	}

	added := AddProducts(db, seller_id, addForProducts)
	updated := UpdateProducts(db, updateForProducts)
	deleted := DeleteProducts(db, seller_id, deleteForProducts)

	declaration := declaration{
		added:   added,
		updated: updated,
		deleted: deleted,
		wrong:   wrong,
	}

	return declaration
}

//глубокая идея с подменой данных
func IsSimilar(dbProduct, excelProduct Product) bool {
	isSimilar := true
	if dbProduct.name != excelProduct.name {
		isSimilar = false
	}
	if dbProduct.price != excelProduct.price {
		isSimilar = false
	}
	return isSimilar
}