// Copyright 2013-2016 Adam Presley. All rights reserved
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package mailslurper

import (
	"fmt"
	"log"
	"strings"
)

type MailHeader struct {
	ContentType string
	Boundary    string
	MIMEVersion string
	Subject     string
	Date        string
	XMailer     string
}

func NewMailHeader(contentType, boundary, mimeVersion, subject, date, xMailer string) *MailHeader {
	return &MailHeader{
		ContentType: contentType,
		Boundary:    boundary,
		MIMEVersion: mimeVersion,
		Subject:     subject,
		Date:        date,
		XMailer:     xMailer,
	}
}

/*
ParseMailHeader, given an entire mail transmission this method parses a set of mail headers.
It will split lines up and figures out what header data goes into what
structure key. Most headers follow this format:

Header-Name: Some value here\r\n

However some headers, such as Content-Type, may have additional information,
especially when the content type is a multipart and there are attachments.
Then it can look like this:

Content-Type: multipart/mixed; boundary="==abcsdfdfd=="\r\n
*/
func (this *MailHeader) Parse(contents string) error {
	var key string

	this.XMailer = "MailSlurper!"
	this.Boundary = ""

	/*
	 * Split the DATA content by CRLF CRLF. The first item will be the data
	 * headers. Everything past that is body/message.
	 */
	headerBodySplit := strings.Split(contents, "\r\n\r\n")
	if len(headerBodySplit) < 2 {
		return fmt.Errorf("Expected DATA block to contain a header section and a body section")
	}

	contents = headerBodySplit[0]

	/*
	 * Unfold and split the header into lines. Loop over each line
	 * and figure out what headers are present. Store them.
	 * Sadly some headers require special processing.
	 */
	contents = UnfoldHeaders(contents)
	splitHeader := strings.Split(contents, "\r\n")
	numLines := len(splitHeader)

	for index := 0; index < numLines; index++ {
		splitItem := strings.Split(splitHeader[index], ":")
		key = splitItem[0]

		switch strings.ToLower(key) {
		case "content-type":
			contentType := strings.Join(splitItem[1:], "")
			contentTypeSplit := strings.Split(contentType, ";")

			this.ContentType = strings.TrimSpace(contentTypeSplit[0])
			log.Println("libmailslurper: INFO - Mail Content-Type: ", this.ContentType)

			/*
			 * Check to see if we have a boundary marker
			 */
			if len(contentTypeSplit) > 1 {
				contentTypeRightSide := strings.Join(contentTypeSplit[1:], ";")

				if strings.Contains(strings.ToLower(contentTypeRightSide), "boundary") {
					boundarySplit := strings.Split(contentTypeRightSide, "=")
					this.Boundary = strings.Replace(strings.Join(boundarySplit[1:], "="), "\"", "", -1)

					log.Println("libmailslurper: INFO - Mail Boundary: ", this.Boundary)
				}
			}

		case "date":
			this.Date = ParseDateTime(strings.Join(splitItem[1:], ":"))
			log.Println("libmailslurper: INFO - Mail Date: ", this.Date)

		case "mime-version":
			this.MIMEVersion = strings.TrimSpace(strings.Join(splitItem[1:], ""))
			log.Println("libmailslurper: INFO - Mail MIME-Version: ", this.MIMEVersion)

		case "subject":
			this.Subject = strings.TrimSpace(strings.Join(splitItem[1:], ""))
			if this.Subject == "" {
				this.Subject = "(No Subject)"
			}

			log.Println("libmailslurper: INFO - Mail Subject: ", this.Subject)
		}
	}

	return nil
}
