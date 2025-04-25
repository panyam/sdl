package views

import "math/rand"

var ADJECTIVES = []string{
	"Pancreatic", "Zesty", "Infinite", "Probabilistic", "Lopsided", "Ambidextrous", "Wobbly", "Turbo-juiced", "Perplexing", "Gloriously Inverted", "Chrono-melty", "Irrational", "Paradoxical", "Fuzzy", "Anti-gravitational", "Quantum-Caffeinated", "Mildly Explosive", "Snark-infused", "Retro-reactive", "Blissfully Chaotic", "Pseudo-Magnetic", "Folded", "Bioluminescent", "Disillusioned", "Meta-sentient", "Vibro-oscillating", "Underwhelming", "Flimsy but Determined", "Jellified", "Emotionally Compromised", "Overclocked", "Pretentious", "Sparkle-Driven", "Sub-neural", "Inappropriately Enthusiastic", "Dystopically Curious", "Partially Sentient", "Sublime", "Unreasonably Confident", "Self-loathing", "Exothermic", "Theoretical", "Extraterrestrially Fashionable", "Existentially Damp", "Chronically Confused", "Planet-Eating", "Singularity-Adjacent", "Reluctantly Vibrating", "Thermally Dubious", "Temporally Ironic",
}

var NOUNS = []string{
	"Chronocube", "Quazmodrive", "Neuronicator", "Glibblebot", "Plasmapod", "Snarkometer", "Hyperplinth", "Zogwheel", "Fluxiglobe", "Thoughtspanner", "Entropulator", "Quantum Crate", "Zeitwobble", "Datawriggler", "Nebuloscope", "Gravatron", "Dimensiplex", "Echo Sphere", "Mechaworm", "Cosmic Combobulator", "Astroduct", "Wibblifier", "Nano-Fling", "Gloopnozzle", "Event Spindle", "Synapsojack", "Pangalactic Prism", "Moodwarp Node", "Starcrank", "Time Funnel", "Reality Siphon", "Ion Chute", "Blipcore", "Turbo-Hoojamaflip", "Subspace Cradle", "Cortex Vortex", "Blink Engine", "Tangentifier", "Planet Noodler", "Relativity Paddle", "Singularity Biscuit", "Wormhole Harness", "Brain Grenade", "Holovessel", "Gravimuffin", "Space Wrench", "Logic Duct", "Meta-Rake", "Anomaly Whisk", "Dreadnut",
}

func randomDesignName() (out string) {
	adj := ADJECTIVES[rand.Intn(len(ADJECTIVES))]
	noun := NOUNS[rand.Intn(len(NOUNS))]
	return adj + " " + noun
}
