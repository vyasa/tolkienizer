package pfa

import (
   "os"
   "fmt"
   "sync"
)

type transitionTable struct {
   table      map[string]*transitionTable
   transition chan string
   mutex      *sync.RWMutex
}

func newTransitionTable() *transitionTable {
   return &transitionTable{
      table:      make(map[string]*transitionTable),
      transition: nil,
      mutex:      new(sync.RWMutex)}
}

// Look up a channel in a transition table.
func (me *transitionTable) lookUp(ss []string) (chan string, os.Error) {

   // If this is the case, we're at the end of our search, just return the next
   // PFA node.
   me.mutex.RLock()
   if len(ss) == 0 {
      defer me.mutex.RUnlock()
      return me.transition, nil
   }

   // Safely lookup the next step in the chain
   t, ok := me.table[ss[0]]
   me.mutex.RUnlock()

   if ok {
      // It exists, recurse!
      return t.lookUp(ss[1:])
   } else {
      // If not, return nil
      return nil, os.NewError("No entry for " + fmt.Sprintf("'%v'.", ss))
   }

   panic("return path bug")
}

// Add a node that doesn't exist to the transition table
func (me *transitionTable) insert(ss []string, cs chan string) os.Error {

   // If this is the case, we are at the node we need to modify
   if len(ss) == 0 {
      // If this node is unmodified, set it to cs
      if me.transition == nil {
         // Make sure we're the only ones touching the table when we alter it
         me.mutex.Lock()
         me.transition = cs
         me.mutex.Unlock()

         // No error, return nil
         return nil
      } else {
         // Tried to assign cs to a node that was already assigned
         return os.NewError("Entry already exists for " + fmt.Sprintf("'%v'.", ss))
      }
   }

   // Looks like we need to look deeper, lock the mutex while we look at the
   // table.
   me.mutex.RLock()
   t, ok := me.table[ss[0]]
   me.mutex.RUnlock()

   if ok {
      // If there was something in the table it's safe to recurse.
      return t.insert(ss[1:], cs)
   } else {
      // Lock the table while we modify it
      me.mutex.Lock()
      // Create a new map entry on the way to our final node
      me.table[ss[0]] = &transitionTable{
         table:      make(map[string]*transitionTable),
         transition: nil,
         mutex:      me.mutex}
      me.mutex.Unlock()

      return t.insert(ss[1:], cs)
   }

   panic("return path bug")
}