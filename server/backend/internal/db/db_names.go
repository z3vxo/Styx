package db

import "math/rand/v2"

// Agent callsigns combine a Greek-myth figure with a Greek-myth place,
// creature, or concept, e.g. "Nyx-Tartarus".
var agentNameFigures = []string{
	"Aether", "Achilles", "Ajax", "Atlas", "Chaos", "Chronos", "Circe",
	"Coeus", "Cronus", "Erebus", "Eris", "Eros", "Gaia", "Hades", "Hector",
	"Helios", "Hemera", "Hera", "Heracles", "Hermes", "Hyperion", "Iapetus",
	"Iris", "Kratos", "Lethe", "Mnemosyne", "Nyx", "Oceanus", "Odysseus",
	"Orpheus", "Pan", "Persephone", "Phoebe", "Pontus", "Poseidon",
	"Prometheus", "Rhea", "Selene", "Styx", "Tartarus", "Thanatos", "Theia",
	"Themis", "Theseus", "Uranus", "Zephyrus", "Zeus",
}

var agentNameSymbols = []string{
	"Acropolis", "Aegis", "Ambrosia", "Asphodel", "Chimera", "Cyclops",
	"Delphi", "Elysium", "Gorgon", "Hydra", "Ichor", "Labyrinth", "Laurel",
	"Lyre", "Minotaur", "Nectar", "Olympus", "Oracle", "Pegasus", "Phalanx",
	"Phoenix", "Siren", "Thunder", "Titan", "Trident", "Underworld", "Veil",
	"Wings",
}

// GenerateAgentName returns a two-word Greek-myth callsign.
func GenerateAgentName() string {
	figure := agentNameFigures[rand.IntN(len(agentNameFigures))]
	symbol := agentNameSymbols[rand.IntN(len(agentNameSymbols))]
	return figure + "-" + symbol
}
