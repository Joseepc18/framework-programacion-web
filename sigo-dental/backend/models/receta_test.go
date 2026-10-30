package models

import "testing"

// Recetar un fármaco al que el paciente es alérgico (o de su misma familia) muestra el aviso.
func TestAvisosDeAlergia(t *testing.T) {
	casos := []struct {
		nombre, alergias string
		medicamentos     []string
		avisos           int
	}{
		{"mismo fármaco", "Penicilina", []string{"Penicilina V 500 mg"}, 1},
		{"misma familia: penicilina y amoxicilina", "Alergia a la penicilina", []string{"Amoxicilina 500 mg", "Paracetamol 500 mg"}, 1},
		{"con acentos y mayúsculas", "CODEÍNA", []string{"codeina 30 mg"}, 1},
		{"familia de los AINE", "AINEs", []string{"Ibuprofeno 400 mg", "Diclofenaco"}, 2},
		{"alergia a un AINE y receta de otro", "Ibuprofeno", []string{"Naproxeno 550 mg"}, 1},
		{"sin alergias declaradas", "Ninguna", []string{"Amoxicilina", "Ibuprofeno"}, 0},
		{"niega alergias", "Niega alergias conocidas", []string{"Amoxicilina"}, 0},
		{"alergia distinta", "Látex y polen", []string{"Amoxicilina"}, 0},
		{"sin medicamentos", "Penicilina", nil, 0},
		{"varias alergias", "Penicilina, sulfas", []string{"Amoxicilina", "Sulfametoxazol", "Paracetamol"}, 2},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if avisos := AvisosDeAlergia(c.alergias, c.medicamentos); len(avisos) != c.avisos {
				t.Errorf("avisos = %v, se esperaban %d", avisos, c.avisos)
			}
		})
	}
}
