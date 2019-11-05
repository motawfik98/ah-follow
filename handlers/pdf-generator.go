package handlers

import (
	"bytes"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"html/template"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

//pdf requestpdf struct
type RequestPdf struct {
	body string
}

//new request to pdf function
func NewRequestPdf(body string) *RequestPdf {
	return &RequestPdf{
		body: body,
	}
}

//parsing template function
func (r *RequestPdf) ParseTemplate(templateFileName string, data interface{}) error {

	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	r.body = buf.String()
	return nil
}

//generate pdf function
func (r *RequestPdf) GeneratePDF() (*wkhtmltopdf.PDFGenerator, error) {
	t := time.Now().Unix()
	// write whole the body
	fileName := "final-reports/" + strconv.FormatInt(int64(t), 10) + ".html"
	err := ioutil.WriteFile(fileName, []byte(r.body), 0644)
	if err != nil {
		panic(err)
	}

	f, err := os.Open(fileName)
	if f != nil {
		defer f.Close()
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		panic(err)
	}
	pdfg.Orientation.Set(wkhtmltopdf.OrientationLandscape)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	//pdfg.Grayscale.Set(true)

	page := wkhtmltopdf.NewPageReader(f)

	page.FooterRight.Set("[page]/[toPage]")
	page.FooterFontName.Set("Calibri")

	//page.FooterFontSize.Set(10)

	page.HeaderLeft.Set(time.Now().Format("02/01/2006"))

	pdfg.AddPage(page)

	pdfg.Dpi.Set(300)

	err = pdfg.Create()
	if err != nil {
		panic(err)
	}

	return pdfg, nil
}
