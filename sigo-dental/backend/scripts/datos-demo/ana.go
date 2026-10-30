package main

import (
	"fmt"
)

// cargarAna completa el expediente de Ana Prueba (paciente 1), la primera de la lista: alérgica a la penicilina y
// anticoagulada. Ya tiene citas, un signo vital, una receta y un cobro; aquí se añade lo que le falta para que todas
// sus pestañas tengan datos. No toca a los demás pacientes.
func (d *demo) cargarAna(c *cliente) {
	ana := d.p(1)
	var cita int
	hoy := d.hoy.Format("2006-01-02")
	// Los huecos de Ana empiezan a las 15:00 para no chocar con los de la primera carga, sea cual sea el día.
	d.hora = 14

	fmt.Println("\nExpediente completo de Ana Prueba")
	paso("Ana: cita de la visita (en atención)", func() (string, error) {
		var err error
		cita, err = d.cita(c, ana, d.proc[3], "Revisión y tratamiento", est("confirmado"), est("en_sala"), est("en_atencion"))
		return fmt.Sprintf("cita %d", cita), err
	})
	paso("Ana: archivo", func() (string, error) {
		return "1 radiografía", subir(c, ana, "radiografia", "Radiografía periapical de ejemplo", "periapical-ana-demo.png", pngMinimo)
	})

	var caries, sellante int
	paso("Ana: odontograma (requeridos, existentes y realizados)", func() (string, error) {
		hallazgos := []mapa{
			{"pieza": 26, "superficie": "oclusal", "tipo": "caries", "estado": "requerido", "observacion": "Caries media"},
			{"pieza": 36, "superficie": "oclusal", "tipo": "sellante", "estado": "requerido"},
			{"pieza": 25, "tipo": "implante", "estado": "requerido", "observacion": "Valorar tras control con su médico"},
			{"pieza": 15, "superficie": "distal", "tipo": "restauracion", "estado": "existente"},
			{"pieza": 38, "tipo": "ausente", "estado": "existente"},
			{"pieza": 46, "tipo": "endodoncia", "estado": "realizado", "observacion": "Tratamiento previo en la clínica"},
			{"pieza": 11, "tipo": "protesis_fija", "estado": "existente"},
		}
		for _, h := range hallazgos {
			var creado id
			if err := c.enviar("odontologo", "POST", ruta(ana, "/odontograma/hallazgos"), h, &creado); err != nil {
				return "", err
			}
			switch h["tipo"] {
			case "caries":
				caries = creado.ID
			case "sellante":
				sellante = creado.ID
			}
		}
		// Segunda parte de la visita: se trata la caries y se sella, y el historial conserva lo anterior.
		if caries != 0 {
			if err := c.enviar("odontologo", "POST", ruta(ana, fmt.Sprintf("/odontograma/hallazgos/%d/retirar", caries)), nil, nil); err != nil {
				return "", err
			}
			if err := c.enviar("odontologo", "POST", ruta(ana, "/odontograma/hallazgos"), mapa{"pieza": 26, "superficie": "oclusal", "tipo": "restauracion", "estado": "realizado", "observacion": "Resina compuesta"}, nil); err != nil {
				return "", err
			}
		}
		if sellante != 0 {
			if err := c.enviar("odontologo", "PUT", ruta(ana, fmt.Sprintf("/odontograma/hallazgos/%d", sellante)), mapa{"pieza": 36, "superficie": "oclusal", "tipo": "sellante", "estado": "realizado"}, nil); err != nil {
				return "", err
			}
		}
		return "7 hallazgos; la caries se retiró y se realizaron la restauración y el sellante", nil
	})
	paso("Ana: periodontograma (dos exámenes)", func() (string, error) {
		// Con la anticoagulación el sangrado al sondaje es mayor: se marca en más puntos.
		if err := d.examen(c, ana, d.hoy.AddDate(0, 0, -45).Format("2006-01-02"), [6]int{4, 5, 4, 4, 4, 3}, [6]int{1, 2, 1, 1, 1, 0}, [6]bool{true, true, true, false, true, true}, 2); err != nil {
			return "", err
		}
		return "exámenes de hace 45 días y de hoy", d.examen(c, ana, hoy, [6]int{3, 3, 3, 3, 2, 2}, [6]int{1, 1, 0, 1, 0, 0}, [6]bool{false, true, false, false, true, false}, 1)
	})
	paso("Ana: evolución SOEP firmada y adenda", func() (string, error) {
		var nota id
		err := c.enviar("odontologo", "POST", ruta(ana, "/notas"), mapa{
			"cita_id":    cita,
			"subjetivo":  marca + " Refiere sangrado de encías al cepillarse y sensibilidad en el molar superior izquierdo.",
			"objetivo":   "Sangrado al sondaje generalizado, caries oclusal en 26. Paciente anticoagulada: se confirmó la dosis con su médico antes de intervenir.",
			"evaluacion": "Caries media en 26 y gingivitis asociada a la anticoagulación.",
			"plan":       "Restaurar 26 con hemostasia local, sellar 36, control periodontal. Evitar penicilinas (alergia): usar clindamicina si hace falta antibiótico.",
		}, &nota)
		if err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(ana, fmt.Sprintf("/notas/%d/firmar", nota.ID)), mapa{"password": c.password}, nil); err != nil {
			return "", err
		}
		return "nota firmada y adenda", c.enviar("odontologo", "POST", ruta(ana, "/adendas"), mapa{
			"nota_id": nota.ID, "texto": "Aclaración: el control de coagulación (INR) del día fue 2,4, dentro del rango aceptado para el procedimiento.", "password": c.password,
		}, nil)
	})
	paso("Ana: consentimiento firmado con OTP y procedimiento", func() (string, error) {
		var k id
		if err := c.enviar("odontologo", "POST", ruta(ana, "/consentimientos"), mapa{"procedimiento_id": d.proc[4], "pieza": 28}, &k); err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(ana, fmt.Sprintf("/consentimientos/%d/otp", k.ID)), nil, nil); err != nil {
			return "", err
		}
		codigo, err := codigoDe(c, ana, "otp_consentimiento")
		if err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(ana, fmt.Sprintf("/consentimientos/%d/firmar", k.ID)), mapa{"modalidad": "otp", "codigo": codigo, "dispositivo": "script de demostración"}, nil); err != nil {
			return "", err
		}
		var p id
		if err := c.enviar("odontologo", "POST", ruta(ana, "/procedimientos-realizados"), mapa{"procedimiento_id": d.proc[3], "pieza": 26, "superficie": "oclusal", "notas": "Restauración con hemostasia local"}, &p); err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(ana, fmt.Sprintf("/procedimientos-realizados/%d/completar", p.ID)), nil, nil); err != nil {
			return "", err
		}
		return fmt.Sprintf("consentimiento %d (extracción, pieza 28) firmado y una restauración realizada", k.ID), nil
	})
	paso("Ana: cierre de la visita", func() (string, error) {
		return "cita atendida", c.enviar("recepcion", "POST", fmt.Sprintf("/api/citas/%d/estado", cita), mapa{"estado": "atendido"}, nil)
	})
	paso("Ana: presupuesto aprobado con una tarea agendada", func() (string, error) {
		var p struct {
			ID    int  `json:"id"`
			Items []id `json:"items"`
		}
		err := c.enviar("odontologo", "POST", ruta(ana, "/presupuestos"), mapa{
			"notas": "Plan de mantenimiento periodontal y rehabilitación",
			"items": []mapa{
				{"fase": 1, "sesion": 1, "procedimiento_id": d.proc[2]},
				{"fase": 1, "sesion": 2, "procedimiento_id": d.proc[3], "pieza": 15, "superficie": "distal", "descuento_centavos": 500},
				{"fase": 2, "sesion": 1, "procedimiento_id": d.proc[5], "pieza": 16},
			},
			"autorizacion": mapa{"email": cuentas["odontologo"], "password": c.password},
		}, &p)
		if err != nil {
			return "", err
		}
		if len(p.Items) < 3 {
			return "", fmt.Errorf("el presupuesto no devolvió sus ítems")
		}
		if err := c.enviar("recepcion", "POST", ruta(ana, fmt.Sprintf("/presupuestos/%d/presentar", p.ID)), nil, nil); err != nil {
			return "", err
		}
		if err := c.enviar("recepcion", "POST", ruta(ana, fmt.Sprintf("/presupuestos/%d/resolver", p.ID)), mapa{"aprobados": []int{p.Items[0].ID, p.Items[1].ID, p.Items[2].ID}}, nil); err != nil {
			return "", err
		}
		inicio, fin := d.hueco()
		var agendada id
		if err := c.enviar("recepcion", "POST", "/api/citas", mapa{
			"paciente_id": ana, "profesional_id": d.prof, "sillon_id": d.sillon, "procedimiento_id": d.proc[2],
			"motivo": "Limpieza aprobada en el presupuesto", "inicio": inicio, "fin": fin, "presupuesto_item_id": p.Items[0].ID,
		}, &agendada); err != nil {
			return "", err
		}
		return fmt.Sprintf("aprobado total (3 ítems, uno con descuento autorizado) y la limpieza agendada en la cita %d", agendada.ID), nil
	})
}
