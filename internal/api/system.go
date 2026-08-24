package api

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/wyw14/cry-110/internal/access"
	"github.com/wyw14/cry-110/internal/brake"
	"github.com/wyw14/cry-110/internal/crane"
	"github.com/wyw14/cry-110/internal/fire"
	"github.com/wyw14/cry-110/internal/gearbox"
	"github.com/wyw14/cry-110/internal/generator"
	"github.com/wyw14/cry-110/internal/hydraulic"
	"github.com/wyw14/cry-110/internal/icing"
	"github.com/wyw14/cry-110/internal/interlock"
	"github.com/wyw14/cry-110/internal/journal"
	"github.com/wyw14/cry-110/internal/lightning"
	"github.com/wyw14/cry-110/internal/pitch"
	"github.com/wyw14/cry-110/internal/rotor"
	"github.com/wyw14/cry-110/internal/turninggear"
	"github.com/wyw14/cry-110/internal/ventilation"
	"github.com/wyw14/cry-110/internal/yaw"
)

type System struct {
	Isolation       *IsolationOrchestrator
	Coordinator     *interlock.Coordinator
	Rotor           *rotor.Service
	Torsion         *gearbox.TorsionAnalyzer
	Pitch           *pitch.Service
	BrakeSequence   *brake.Sequence
	Yaw             *yaw.Service
	YawBrakeService *brake.Service
	YawBrake        *yaw.BrakeController
	TurningGear     *turninggear.Service
	TurningPermit   *turninggear.Permit
	MainShaftBrake  *brake.Release
	RotorPermit     *interlock.RotorPermit
	Warmup          *gearbox.Warmup
	Lube            *hydraulic.LubeFlow
	Crane           *crane.Slew
	CranePark       *crane.Park
	Reservations    *interlock.ReservationBook
	Hydraulic       *hydraulic.Service
	Generator       *generator.Service
	Ventilation     *ventilation.CoolingService
	Fire            *fire.Response
	Icing           *icing.Response
	Lightning       *lightning.Trip
	Access          *access.Service
	Journal         *journal.Store
	Snapshots       *journal.SnapshotStore
}

func NewSystem(dataDirectory string, now time.Time) (*System, error) {
	if dataDirectory == "" {
		return nil, fmt.Errorf("data directory is required")
	}
	eventStore, err := journal.Open(filepath.Join(dataDirectory, "events.jsonl"))
	if err != nil {
		return nil, err
	}
	snapshots, err := journal.OpenSnapshots(filepath.Join(dataDirectory, "snapshot.json"))
	if err != nil {
		return nil, err
	}

	book := interlock.NewReservationBook(nil)
	craneSlew := crane.NewSlew(0, 5.5, book)
	yawPlanner := yaw.NewPlanner(0, book)
	thermal := brake.NewThermalModel(220, 20, 15*time.Minute, now)
	yawLimit := interlock.NewYawLimit(0.8)
	brakeService := brake.NewService(thermal, yawLimit)
	yawService := yaw.NewService(yawPlanner, brakeService, yawLimit)
	yawBrake := yaw.NewBrakeController(brakeService)
	cranePark := crane.NewPark()

	pinPermit := interlock.NewPinPermit(2 * time.Second)
	stillness := rotor.NewStillnessObserver(rotor.StillnessConfig{
		MaximumAverageRPM: 0.05, MaximumInstantRPM: 0.08,
		MaximumTorqueNm: 150, RequiredDuration: 2 * time.Second, MaximumGap: 500 * time.Millisecond,
	})
	rotorService := rotor.NewService(stillness, pinPermit)
	torsion := gearbox.NewTorsionAnalyzer(150, 0.08, 2*time.Second)
	aero := rotor.NewAeroController()
	pitchCoordinator := pitch.NewCoordinator(86, 120, 2*time.Second, aero)
	actuator1, err := pitch.NewActuator(1, 0)
	if err != nil {
		return nil, err
	}
	actuator2, err := pitch.NewActuator(2, 0)
	if err != nil {
		return nil, err
	}
	actuator3, err := pitch.NewActuator(3, 0)
	if err != nil {
		return nil, err
	}
	pitchService := pitch.NewService(pitchCoordinator, actuator1, actuator2, actuator3)
	energy := pitch.NewEnergyBudget(1.5)
	for blade := 1; blade <= 3; blade++ {
		if err := energy.SetDemand(pitch.BladeDemand{Blade: blade, RemainingAngle: 90, LitersPerDegree: 0.025}); err != nil {
			return nil, err
		}
	}
	runPermit := interlock.NewRunPermit(140)
	hydraulicService := hydraulic.NewService(hydraulic.Accumulator{
		NominalLiters: 16, PrechargeBar: 100, HeaderBar: 220, OilLiters: 12,
	}, 140, energy, runPermit)

	warmup := gearbox.NewWarmup(18, 2.5, 5)
	if err := warmup.Start("startup"); err != nil {
		return nil, err
	}
	lube := &hydraulic.LubeFlow{}
	if err := lube.Observe(3, 7, 20, now); err != nil {
		return nil, err
	}
	warmup.Observe(lube.WarmupProof("startup"))
	phase := gearbox.NewPhaseService(&rotor.Encoder{}, 0)
	phase.Reset(0)
	controller := turninggear.NewController(phase, 0.5, 0.08)
	turningPermit := turninggear.NewPermit(warmup, 5*time.Second)
	rotorPermit := interlock.NewRotorPermit(2 * time.Second)
	disengage := turninggear.NewDisengage(1.2, rotorPermit)
	turningService := turninggear.NewService(controller, turningPermit, disengage)
	mainShaftBrake := brake.NewRelease(rotorPermit)

	exclusion := interlock.NewExclusion()
	testCycle := rotor.NewTestCycle(exclusion)
	icingResponse := icing.NewResponse(exclusion, testCycle)
	accessService := access.NewService(exclusion)
	ventilationService := ventilation.NewCoolingService()
	fireResponse := fire.NewResponse(ventilationService)
	lightningTrip := lightning.NewTrip(yawBrake, cranePark)

	system := &System{
		Coordinator: interlock.NewCoordinator(), Rotor: rotorService, Torsion: torsion,
		Pitch: pitchService, BrakeSequence: brake.NewSequence(), Yaw: yawService,
		YawBrakeService: brakeService, YawBrake: yawBrake,
		TurningGear: turningService, TurningPermit: turningPermit, MainShaftBrake: mainShaftBrake, RotorPermit: rotorPermit,
		Warmup: warmup, Lube: lube, Crane: craneSlew, CranePark: cranePark, Reservations: book,
		Hydraulic: hydraulicService, Generator: generator.NewService(hydraulicService),
		Ventilation: ventilationService, Fire: fireResponse, Icing: icingResponse,
		Lightning: lightningTrip, Access: accessService, Journal: eventStore, Snapshots: snapshots,
	}
	system.Isolation = NewIsolationOrchestrator(system)
	return system, nil
}

func (s *System) MainShaftBrakeReason() string {
	return s.RotorPermit.BlockReason()
}
