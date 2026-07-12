// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Bill Fassinou
//
// StrongSwan Manager — coeur Community, sous licence AGPL-3.0.
// Les modules premium sont distribues separement sous licence commerciale.

package domain

import "strings"

// WeakAlgos liste les algorithmes cryptographiques déconseillés (cf. app.html / EF-10).
var WeakAlgos = []string{"3des", "des", "md5", "modp1024", "modp768"}

// ScoreResult porte le score de sécurité d'un tunnel et les constats associés.
type ScoreResult struct {
	Score int      `json:"score"`
	Notes []string `json:"notes"`
}

// ScoreTunnel calcule un score de sécurité [5..100] pour un tunnel.
// Port fidèle de la fonction scoreTunnel de la maquette front (app.html) :
//
//	IKEv1 −42 ; 3DES/DES −28 ; MD5 −18 ; modp1024 −16 ; sans PFS −10 ; sans ML-KEM −6.
func ScoreTunnel(t *Tunnel) ScoreResult {
	s := 100
	var notes []string
	j := strings.ToLower(strings.Join(append(append([]string{}, t.ProposalsIKE...), t.ProposalsESP...), " "))

	if t.IKEVersion == 1 {
		s -= 42
		notes = append(notes, "IKEv1 obsolète — migrer vers IKEv2")
	}
	if strings.Contains(j, "3des") || containsWord(j, "des") {
		s -= 28
		notes = append(notes, "Chiffrement 3DES/DES faible")
	}
	if strings.Contains(j, "md5") {
		s -= 18
		notes = append(notes, "Empreinte MD5 faible")
	}
	if strings.Contains(j, "modp1024") || strings.Contains(j, "modp768") {
		s -= 16
		notes = append(notes, "Groupe Diffie-Hellman faible (modp1024)")
	}
	if !t.PFS {
		s -= 10
		notes = append(notes, "Perfect Forward Secrecy désactivé")
	}
	if !strings.Contains(j, "mlkem") {
		s -= 6
		notes = append(notes, "Pas de préparation post-quantique (ML-KEM)")
	}

	if s < 5 {
		s = 5
	}
	if s > 100 {
		s = 100
	}
	return ScoreResult{Score: s, Notes: notes}
}

// containsWord détecte "des" comme mot isolé (et non comme sous-chaîne de "3des"/"aes...").
func containsWord(haystack, word string) bool {
	for _, tok := range strings.FieldsFunc(haystack, func(r rune) bool {
		return r == ' ' || r == '-' || r == '_'
	}) {
		if tok == word {
			return true
		}
	}
	return false
}
