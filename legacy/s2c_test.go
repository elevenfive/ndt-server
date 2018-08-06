package legacy

import (
	"log"
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.LUTC | log.Lshortfile)
}

/*
func TestResponder_ManageS2CTest(t *testing.T) {
	certFile := "cert.pem"
	keyFile := "key.pem"

	// Create key & self-signed certificate.
	err := pipe.Run(
		pipe.Script("Create private key and self-signed certificate",
			pipe.Exec("openssl", "genrsa", "-out", "key.pem"),
			pipe.Exec("openssl", "req", "-new", "-x509", "-key", "key.pem", "-out",
				"cert.pem", "-days", "2", "-subj",
				"/C=XX/ST=State/L=Locality/O=Org/OU=Unit/CN=localhost/emailAddress=test@email.address"),
		),
	)
	if err != nil {
		t.Fatalf("Failed to generate server key and certs: %s", err)
	}

	rootCAs := x509.NewCertPool()
	certs, err := ioutil.ReadFile(certFile)
	if err != nil {
		t.Fatalf("Failed to append %q to RootCAs: %v", certFile, err)
	}
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		t.Logf("No certs appended, using system certs only")
	}

	tests := []struct {
		name     string
		result   chan float64
		duration time.Duration
		want     float64
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name:     "try",
			result:   make(chan float64),
			duration: 20 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Responder{
				result:   tt.result,
				duration: tt.duration,
				certFile: certFile,
				keyFile:  keyFile,
			}
			// Setup stub server to start S2C test.
			mux := http.NewServeMux()
			mux.Handle("/ndt_protocol", http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					upgrader := MakeNdtUpgrader([]string{"ndt"})
					ws, err := upgrader.Upgrade(w, r, nil)
					if err != nil {
						log.Println("ERROR SERVER:", err)
						return
					}
					defer ws.Close()

					got, err := tr.ManageS2CTest(ws, http.HandlerFunc(tr.S2CTestHandler))
					fmt.Println(got, err)
					/*fmt.Println("tt", tt)
					if (err != nil) != tt.wantErr {
						t.Errorf("Responder.ManageS2CTest() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
					if got != tt.want {
						t.Errorf("Responder.ManageS2CTest() = %v, want %v", got, tt.want)
					}

				}))
			// Setup client connection to test server.
			ts := httptest.NewServer(mux)
			defer ts.Close()
			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			dialer := &websocket.Dialer{
				Subprotocols: []string{"ndt"},
				//TLSClientConfig: &tls.Config{
				//RootCAs: rootCAs,
				//},
			}
			t.Logf("dialing control channel")
			ws, _, err := dialer.Dial("ws://localhost:"+u.Port()+"/ndt_protocol", nil)
			if err != nil {
				t.Fatal(err)
			}
			defer ws.Close()

			t.Logf("recv test prepare msg")
			msg, err := RecvNdtJSONMessage(ws, ndt.TestPrepare)
			if err != nil {
				t.Fatal(err)
			}
			tester := &websocket.Dialer{
				Subprotocols: []string{"ndt"},
				TLSClientConfig: &tls.Config{
					RootCAs: rootCAs,
				},
			}
			t.Logf("dial test port")
			s2c, _, err := tester.Dial("wss://localhost:"+msg.Msg+"/ndt_protocol", nil)
			if err != nil {
				t.Fatal(err)
			}
			defer s2c.Close()
			t.Logf("recv test start msg")
			_, err = RecvNdtJSONMessage(ws, ndt.TestStart)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("fin")

		})
	}
}

*/
