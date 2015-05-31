package minifs

import (
  "net/http"
  "crypto/sha1"
  "encoding/hex"
  "os"
  "io"
  "path/filepath"
  "stathat.com/c/consistent"
  "time"
  "encoding/json"
  "math/rand"
  "log"
  "sync"
)

type ServerState struct {
  Peers map[string] int64 `json:"peers"`
  peerList []string `json:-`
}

type Server struct {
  Port string
  Prefix string
  State *ServerState
  Mutex sync.Mutex
  HTTPClient *http.Client
  Cons *consistent.Consistent
}

func (s *Server) Inspect(peer string) {
  log.Printf("Inspecting https://%s/state.json...", peer)
  
  resp, err := s.HTTPClient.Get("https://" + peer + "/state.json")
  
  if err != nil {
    log.Print(err)
    return
  }
  
  frgn := new(ServerState)
  
  if err := json.NewDecoder(resp.Body).Decode(&frgn); err != nil {
    log.Print(err)
    return
  }
  
  for key, age := range frgn.Peers {
    if age > s.State.Peers[key] {
      s.State.Peers[key] = age
    }
  }
  
  if s.Cons != nil {
    present := make([]string, 0)
    
    for peer, age := range s.State.Peers {
      if time.Now().Unix() - age < 30 {
        present = append(present, peer)
      }
    }
    
    s.Cons.Set(present)
  }
}

func (s *Server) RunInspection() {
  if len(s.State.Peers) == 0 {
    return
  }
  
  s.Mutex.Lock()
  defer s.Mutex.Unlock()
  
  sel := rand.Intn(len(s.State.Peers))
  idx := 0
  
  for peer := range s.State.Peers {
    if idx == sel {
      s.Inspect(peer)
    }
    
    idx += 1
  }
}

func (s *Server) isAuthorized(r *http.Request) bool {
  return true // return len(r.TLS.PeerCertificates) > 0
}

func (s *Server) resolve(path string) string {
  sha1 := sha1.New()
  io.WriteString(sha1, path)
  result := hex.EncodeToString(sha1.Sum(nil))

  return s.Prefix + "/" + result[0:2] + "/" + result[2:]
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
  host, err := os.Hostname()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  
  if r.URL.Path == "/state.json" {
    state := s.State
    
    state.Peers[host + s.Port] = time.Now().Unix()
    
    json.NewEncoder(w).Encode(state)
    return
  }

  f, err := os.OpenFile(s.resolve(r.URL.Path), os.O_RDWR, 0666)
  defer f.Close()
  
  if os.IsNotExist(err) {
    peer, err := s.Cons.Get(r.URL.Path)
    
    if err == nil && peer != r.Host {
      http.Redirect(w, r, "https://" + peer + r.URL.Path, 301)
      return
    }
    
    http.NotFound(w, r)

  } else if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  io.Copy(w, f)
}

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path == "/state.json" {
    http.Error(w, "not allowed", http.StatusMethodNotAllowed)
    return
  }
  
  rsv := s.resolve(r.URL.Path)
  if err := os.MkdirAll(filepath.Dir(rsv), 0755); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  f, err := os.OpenFile(rsv, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0600)
  defer f.Close()
  
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  io.Copy(f, r.Body)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  if !s.isAuthorized(r) {
    http.Error(w, "403 Forbidden", http.StatusForbidden)
    return
  }
  
  if r.Method == "PUT" || r.Method == "POST" {
    s.handlePut(w, r)
  } else if r.Method == "GET" {
    s.handleGet(w, r)
  } else {
    http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
  }
}
