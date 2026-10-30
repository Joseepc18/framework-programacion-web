package controllers

import (
	"testing"

	"sigo-dental-backend/models"
)

// Si el paciente es menor de edad firma el tutor legal; sin tutor registrado no se puede preparar el consentimiento.
func TestFirmanteDelConsentimientoSegunLaEdad(t *testing.T) {
	adulto := models.PacienteDetalle{Paciente: models.Paciente{Nombres: "Ana", Apellidos: "Pérez", Documento: "0911", Edad: 30}}
	if tipo, nombre, doc, motivo := firmanteDe(adulto); tipo != models.FirmantePaciente || nombre != "Ana Pérez" || doc != "0911" || motivo != "" {
		t.Errorf("adulto: %s %s %s %q", tipo, nombre, doc, motivo)
	}
	menor := models.PacienteDetalle{Paciente: models.Paciente{Nombres: "Sofía", Apellidos: "Lara", Edad: 10, TutorNombre: "Rosa Lara", TutorDocumento: "0922"}}
	if tipo, nombre, doc, motivo := firmanteDe(menor); tipo != models.FirmanteTutor || nombre != "Rosa Lara" || doc != "0922" || motivo != "" {
		t.Errorf("menor con tutor: %s %s %s %q", tipo, nombre, doc, motivo)
	}
	menor.TutorNombre = ""
	if _, _, _, motivo := firmanteDe(menor); motivo == "" {
		t.Error("un menor sin tutor registrado no debe poder firmar")
	}
}

func TestElCodigoDeUnMenorVaAlTelefonoDelTutor(t *testing.T) {
	menor := models.PacienteDetalle{Paciente: models.Paciente{Telefono: "0991111111", Email: "p@x.test", CanalPreferido: "email", TutorTelefono: "0982222222"}}
	if canal, destino := destinoDelCodigo(menor, models.FirmanteTutor); canal != "sms" || destino != "0982222222" {
		t.Errorf("tutor: %s %s", canal, destino)
	}
	if canal, destino := destinoDelCodigo(menor, models.FirmantePaciente); canal != "email" || destino != "p@x.test" {
		t.Errorf("paciente con correo preferido: %s %s", canal, destino)
	}
}
