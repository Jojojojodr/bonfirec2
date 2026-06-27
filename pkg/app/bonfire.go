package app

import (
	"log"
	"os"

	"github.com/Jojojojodr/bonfirec2"

	"gorm.io/gorm"
)

func RestartServices(db *gorm.DB) {
	var listeners []bonfirec2.Listener
	if db.Find(&listeners) != nil {
		for i := range listeners {
			listener := &listeners[i]
			listener.ReinitializeRuntimeChannels()
			log.Printf("Restarting listener: %s on %s:%s (%s)", listener.ID, listener.Address, listener.Port, listener.Protocol)
			go func(l *bonfirec2.Listener) {
				err := l.Start()
				if err != nil {
					log.Printf("Failed to restart listener %s: %v", l.ID, err)
				}
			}(listener)
			bonfirec2.Listeners[listener.ID] = listener
			log.Printf("Listener %s restarted successfully", listener.ID)
		}
	}

	var grunts []bonfirec2.Grunt
	if db.Find(&grunts) != nil {
		for i := range grunts {
			grunt := &grunts[i]
			log.Printf("Restarting grunt: %s (%s)", grunt.ID, grunt.Address)
			bonfirec2.Grunts[grunt.ID] = grunt
			if listener, ok := bonfirec2.Listeners[grunt.ListenerID]; ok {
				listener.GruntCount++
			}
			log.Printf("Grunt %s restarted successfully", grunt.ID)
		}
	}

	log.Println("All services restarted successfully")
}

func ShutdownServices(sigCh <-chan os.Signal) {
	sig := <-sigCh
	log.Printf("Received signal %s, marking grunts inactive...", sig)
	bonfirec2.SetAllGruntsInactive()
	os.Exit(0)
}
