package main

import (
	"fmt"
	"time"
)

// cargar recorre las funcionalidades de las Fases 1 y 2 con la API. La historia clínica principal es la de Carlos
// Mendoza (paciente 2): dos visitas, tratamientos, consentimientos, recetas, presupuesto y cobro. Ana (1), Sofía (3,
// menor de edad) y Jorge (4) cubren el resto de los estados y casos.
func (d *demo) cargar(c *cliente) {
	ana, carlos, sofia, jorge := d.p(1), d.p(2), d.p(3), d.p(4)
	var citaV1, citaV2 int
	var cobrable int
	hoy := d.hoy.Format("2006-01-02")

	fmt.Println("\nFase 1: agenda y estados de cita")
	paso("citas en todos los estados", func() (string, error) {
		casos := []struct {
			paciente int
			motivo   string
			estados  [][2]string
		}{
			{jorge, "Control (programado)", nil},
			{sofia, "Limpieza (confirmado)", [][2]string{est("confirmado")}},
			{ana, "Consulta (en sala de espera)", [][2]string{est("confirmado"), est("en_sala")}},
			{jorge, "Dolor molar (en atención)", [][2]string{est("en_sala"), est("en_atencion")}},
			{ana, "Revisión (cancelada)", [][2]string{{"cancelado", "El paciente avisó que viaja"}}},
			{sofia, "Control (no asistió)", [][2]string{est("no_asistio")}},
		}
		for _, k := range casos {
			if _, err := d.cita(c, k.paciente, 0, k.motivo, k.estados...); err != nil {
				return "", err
			}
		}
		return "7 estados en total (con las citas existentes: programado, confirmado, en sala, en atención, cancelado y no asistió)", nil
	})

	fmt.Println("\nFase 2: primera visita de Carlos Mendoza")
	paso("cita de la visita 1 (en atención)", func() (string, error) {
		var err error
		citaV1, err = d.cita(c, carlos, d.proc[3], "Revisión general", est("confirmado"), est("en_sala"), est("en_atencion"))
		return fmt.Sprintf("cita %d", citaV1), err
	})
	paso("signos vitales", func() (string, error) {
		tomas := []struct {
			paciente int
			datos    mapa
		}{
			{carlos, mapa{"sistolica": 118, "diastolica": 76, "frecuencia_cardiaca": 72, "frecuencia_respiratoria": 16, "temperatura": 36.6, "saturacion": 98, "peso_kg": 78.5, "talla_cm": 175}},
			{ana, mapa{"sistolica": 122, "diastolica": 80, "frecuencia_cardiaca": 68, "temperatura": 36.4, "saturacion": 99}},
			{sofia, mapa{"temperatura": 36.8, "peso_kg": 28.4, "talla_cm": 130, "observaciones": "Control de niño sano"}},
			{jorge, mapa{"sistolica": 138, "diastolica": 88, "frecuencia_cardiaca": 80, "saturacion": 96}},
		}
		for _, t := range tomas {
			if err := c.enviar("asistente", "POST", ruta(t.paciente, "/signos-vitales"), t.datos, nil); err != nil {
				return "", err
			}
		}
		return "4 tomas (asistente)", nil
	})
	paso("archivos del paciente", func() (string, error) {
		if err := subir(c, carlos, "radiografia", "Panorámica de ingreso (imagen de ejemplo)", "panoramica-demo.png", pngMinimo); err != nil {
			return "", err
		}
		return "1 radiografía y 1 PDF", subir(c, carlos, "documento", "Informe de laboratorio (PDF de ejemplo)", "informe-demo.pdf", []byte(pdfMinimo))
	})

	var caries, sellante int
	paso("odontograma (visita 1)", func() (string, error) {
		hallazgos := []mapa{
			{"pieza": 16, "superficie": "oclusal", "tipo": "caries", "estado": "requerido", "observacion": "Caries media"},
			{"pieza": 26, "superficie": "oclusal", "tipo": "sellante", "estado": "requerido"},
			{"pieza": 18, "tipo": "exodoncia", "estado": "requerido", "observacion": "Tercer molar incluido"},
			{"pieza": 21, "superficie": "mesial", "tipo": "restauracion", "estado": "existente"},
			{"pieza": 48, "tipo": "ausente", "estado": "existente"},
			{"pieza": 36, "tipo": "endodoncia", "estado": "realizado", "observacion": "Tratamiento previo en la clínica"},
			{"pieza": 46, "tipo": "implante", "estado": "existente"},
			{"pieza": 14, "tipo": "protesis_fija", "estado": "existente"},
		}
		for _, h := range hallazgos {
			var creado id
			if err := c.enviar("odontologo", "POST", ruta(carlos, "/odontograma/hallazgos"), h, &creado); err != nil {
				return "", err
			}
			switch h["tipo"] {
			case "caries":
				caries = creado.ID
			case "sellante":
				sellante = creado.ID
			}
		}
		return fmt.Sprintf("%d hallazgos requeridos, existentes y realizados", len(hallazgos)), nil
	})
	paso("periodontograma (examen inicial)", func() (string, error) {
		return "examen de hace 30 días", d.examen(c, carlos, d.hoy.AddDate(0, 0, -30).Format("2006-01-02"), [6]int{5, 6, 5, 4, 5, 4}, [6]int{2, 3, 2, 1, 2, 1}, [6]bool{true, true, false, true, false, true}, 3)
	})
	paso("evolución SOEP firmada y adenda", func() (string, error) {
		var nota id
		err := c.enviar("odontologo", "POST", ruta(carlos, "/notas"), mapa{
			"cita_id":    citaV1,
			"subjetivo":  marca + " Refiere sensibilidad al frío en el molar superior derecho.",
			"objetivo":   "Caries oclusal en 16, sellante pendiente en 26, sondaje aumentado en 16.",
			"evaluacion": "Caries media en 16 y gingivitis localizada.",
			"plan":       "Restaurar 16, sellar 26, control periodontal en 30 días.",
		}, &nota)
		if err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/notas/%d/firmar", nota.ID)), mapa{"password": c.password}, nil); err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, "/adendas"), mapa{
			"nota_id": nota.ID, "texto": "Aclaración: la sensibilidad es en la pieza 16, no en la 17. Se corrige sin modificar la nota original.", "password": c.password,
		}, nil); err != nil {
			return "", err
		}
		// Una nota del asistente, sin firmar: queda editable hasta que el profesional la firme.
		var perfil []id
		_ = c.enviar("recepcion", "GET", "/api/profesionales", nil, &perfil)
		if len(perfil) > 0 {
			_ = c.enviar("asistente", "POST", ruta(carlos, "/notas"), mapa{
				"profesional_id": perfil[0].ID, "subjetivo": "Paciente colaborador.", "objetivo": "Se tomaron signos vitales y radiografía.",
				"evaluacion": "Sin novedades.", "plan": "Pasa a consulta con el odontólogo.",
			}, nil)
		}
		return "nota firmada (huella SHA-256), adenda firmada y nota del asistente pendiente de firma", nil
	})
	paso("cierre de la visita 1", func() (string, error) {
		return "cita atendida", c.enviar("recepcion", "POST", fmt.Sprintf("/api/citas/%d/estado", citaV1), mapa{"estado": "atendido"}, nil)
	})

	fmt.Println("\nFase 2: consentimientos")
	paso("consentimiento firmado con código OTP", func() (string, error) {
		var k id
		if err := c.enviar("odontologo", "POST", ruta(carlos, "/consentimientos"), mapa{"procedimiento_id": d.proc[4], "pieza": 18}, &k); err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/consentimientos/%d/otp", k.ID)), nil, nil); err != nil {
			return "", err
		}
		codigo, err := codigoDe(c, carlos, "otp_consentimiento")
		if err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/consentimientos/%d/firmar", k.ID)), mapa{"modalidad": "otp", "codigo": codigo, "dispositivo": "script de demostración"}, nil); err != nil {
			return "", err
		}
		return fmt.Sprintf("consentimiento %d (extracción, pieza 18) firmado con OTP; PDF guardado", k.ID), nil
	})
	paso("consentimiento firmado y luego revocado", func() (string, error) {
		var k id
		if err := c.enviar("odontologo", "POST", ruta(carlos, "/consentimientos"), mapa{"procedimiento_id": d.proc[5], "pieza": 36}, &k); err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/consentimientos/%d/firmar", k.ID)), mapa{"modalidad": "manuscrita", "firma": firmaDePrueba(), "dispositivo": "tablet de demostración"}, nil); err != nil {
			return "", err
		}
		var r id
		if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/consentimientos/%d/revocar", k.ID)), mapa{"motivo": "El paciente prefiere esperar y pedir una segunda opinión"}, &r); err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/consentimientos/%d/otp", r.ID)), nil, nil); err != nil {
			return "", err
		}
		codigo, err := codigoDe(c, carlos, "otp_consentimiento")
		if err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/consentimientos/%d/firmar", r.ID)), mapa{"modalidad": "otp", "codigo": codigo}, nil); err != nil {
			return "", err
		}
		return fmt.Sprintf("consentimiento %d (endodoncia) firmado a mano y revocado con el documento %d", k.ID, r.ID), nil
	})
	paso("consentimiento de un menor (firma el tutor)", func() (string, error) {
		return "pendiente de firma del tutor de Sofía", c.enviar("odontologo", "POST", ruta(sofia, "/consentimientos"), mapa{"procedimiento_id": d.proc[4], "pieza": 75}, nil)
	})

	fmt.Println("\nFase 2: segunda visita de Carlos Mendoza")
	paso("cita de la visita 2 (en atención)", func() (string, error) {
		var err error
		citaV2, err = d.cita(c, carlos, d.proc[3], "Tratamiento de la caries", est("confirmado"), est("en_sala"), est("en_atencion"))
		return fmt.Sprintf("cita %d", citaV2), err
	})
	paso("odontograma (visita 2: corregir, retirar y realizar)", func() (string, error) {
		if caries != 0 {
			if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/odontograma/hallazgos/%d/retirar", caries)), nil, nil); err != nil {
				return "", err
			}
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, "/odontograma/hallazgos"), mapa{"pieza": 16, "superficie": "oclusal", "tipo": "restauracion", "estado": "realizado", "observacion": "Resina compuesta"}, nil); err != nil {
			return "", err
		}
		if sellante != 0 {
			if err := c.enviar("odontologo", "PUT", ruta(carlos, fmt.Sprintf("/odontograma/hallazgos/%d", sellante)), mapa{"pieza": 26, "superficie": "oclusal", "tipo": "sellante", "estado": "realizado"}, nil); err != nil {
				return "", err
			}
		}
		return "caries retirada, restauración y sellante realizados (el historial conserva lo anterior)", nil
	})
	paso("procedimientos realizados", func() (string, error) {
		var p id
		if err := c.enviar("odontologo", "POST", ruta(carlos, "/procedimientos-realizados"), mapa{"procedimiento_id": d.proc[3], "pieza": 16, "superficie": "oclusal", "notas": "Restauración de la caries oclusal"}, &p); err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/procedimientos-realizados/%d/completar", p.ID)), nil, nil); err != nil {
			return "", err
		}
		// Con el consentimiento firmado y vigente sí se puede iniciar la extracción.
		var e id
		if err := c.enviar("odontologo", "POST", ruta(carlos, "/procedimientos-realizados"), mapa{"procedimiento_id": d.proc[4], "pieza": 18}, &e); err != nil {
			return "", err
		}
		return "restauración realizada y extracción en curso (permitida por el consentimiento)", nil
	})
	paso("periodontograma (examen de control para comparar)", func() (string, error) {
		return "examen de hoy, con mejoría", d.examen(c, carlos, hoy, [6]int{3, 4, 3, 3, 3, 2}, [6]int{1, 2, 1, 0, 1, 0}, [6]bool{false, true, false, false, false, false}, 2)
	})
	paso("cierre de la visita 2", func() (string, error) {
		cobrable = citaV2
		return "cita atendida", c.enviar("recepcion", "POST", fmt.Sprintf("/api/citas/%d/estado", citaV2), mapa{"estado": "atendido"}, nil)
	})

	fmt.Println("\nFase 2: recetas")
	paso("recetas (firmada, con aviso de alergia y anulada)", func() (string, error) {
		meds := []mapa{
			{"nombre": "Ibuprofeno 400 mg", "dosis": "1 tableta", "frecuencia": "cada 8 horas", "duracion": "3 días", "indicaciones": "Después de las comidas"},
			{"nombre": "Paracetamol 500 mg", "dosis": "1 tableta", "frecuencia": "cada 6 horas si hay dolor", "duracion": "3 días"},
		}
		var r1 id
		if err := c.enviar("odontologo", "POST", ruta(carlos, "/recetas"), mapa{"medicamentos": meds, "indicaciones": "Dieta blanda por 24 horas"}, &r1); err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/recetas/%d/firmar", r1.ID)), mapa{"password": c.password}, nil); err != nil {
			return "", err
		}
		var r2 id
		if err := c.enviar("odontologo", "POST", ruta(carlos, "/recetas"), mapa{"medicamentos": []mapa{{"nombre": "Clindamicina 300 mg", "dosis": "1 cápsula", "frecuencia": "cada 8 horas", "duracion": "7 días"}}}, &r2); err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/recetas/%d/firmar", r2.ID)), mapa{"password": c.password}, nil); err != nil {
			return "", err
		}
		if err := c.enviar("odontologo", "POST", ruta(carlos, fmt.Sprintf("/recetas/%d/anular", r2.ID)), mapa{"motivo": "Dosis mal indicada; se reemplaza por otra", "copiar": true}, nil); err != nil {
			return "", err
		}
		var alerta struct {
			Avisos []mapa `json:"avisos_alergia"`
		}
		if err := c.enviar("odontologo", "POST", ruta(ana, "/recetas"), mapa{"medicamentos": []mapa{{"nombre": "Amoxicilina 500 mg", "dosis": "1 cápsula", "frecuencia": "cada 8 horas", "duracion": "7 días"}}}, &alerta); err != nil {
			return "", err
		}
		if len(alerta.Avisos) == 0 {
			return "", fmt.Errorf("la receta de Ana con amoxicilina debía traer un aviso de alergia a la penicilina")
		}
		return "1 firmada, 1 anulada (con reemplazo en borrador) y 1 de Ana con aviso de alergia a la penicilina", nil
	})

	fmt.Println("\nFase 2: presupuestos y tareas")
	paso("presupuesto aprobado parcialmente con una tarea agendada", func() (string, error) {
		var p struct {
			ID    int  `json:"id"`
			Items []id `json:"items"`
		}
		err := c.enviar("odontologo", "POST", ruta(carlos, "/presupuestos"), mapa{
			"notas": "Plan de rehabilitación en dos fases",
			"items": []mapa{
				{"fase": 1, "sesion": 1, "procedimiento_id": d.proc[2]},
				{"fase": 1, "sesion": 2, "procedimiento_id": d.proc[3], "pieza": 14, "superficie": "oclusal", "descuento_centavos": 500},
				{"fase": 2, "sesion": 1, "procedimiento_id": d.proc[5], "pieza": 36, "laboratorio_centavos": 2500},
			},
			"autorizacion": mapa{"email": cuentas["odontologo"], "password": c.password},
		}, &p)
		if err != nil {
			return "", err
		}
		if len(p.Items) < 3 {
			return "", fmt.Errorf("el presupuesto no devolvió sus ítems")
		}
		if err := c.enviar("recepcion", "POST", ruta(carlos, fmt.Sprintf("/presupuestos/%d/presentar", p.ID)), nil, nil); err != nil {
			return "", err
		}
		if err := c.enviar("recepcion", "POST", ruta(carlos, fmt.Sprintf("/presupuestos/%d/resolver", p.ID)), mapa{"aprobados": []int{p.Items[0].ID, p.Items[1].ID}}, nil); err != nil {
			return "", err
		}
		inicio, fin := d.hueco()
		var cita id
		if err := c.enviar("recepcion", "POST", "/api/citas", mapa{
			"paciente_id": carlos, "profesional_id": d.prof, "sillon_id": d.sillon, "procedimiento_id": d.proc[3],
			"motivo": "Restauración aprobada en el presupuesto", "inicio": inicio, "fin": fin, "presupuesto_item_id": p.Items[1].ID,
		}, &cita); err != nil {
			return "", err
		}
		return fmt.Sprintf("2 de 3 ítems aprobados (con descuento autorizado) y la tarea de la restauración agendada en la cita %d; la endodoncia quedó rechazada", cita.ID), nil
	})
	paso("presupuesto en borrador", func() (string, error) {
		return "borrador para Jorge", c.enviar("odontologo", "POST", ruta(jorge, "/presupuestos"), mapa{
			"items": []mapa{{"fase": 1, "sesion": 1, "procedimiento_id": d.proc[1]}, {"fase": 1, "sesion": 2, "procedimiento_id": d.proc[2]}},
		}, nil)
	})

	fmt.Println("\nFase 1: cobros")
	paso("cobros (pago parcial, pagado completo y anulado)", func() (string, error) {
		var c1, c2, c3 struct {
			ID     int    `json:"id"`
			Numero string `json:"numero"`
		}
		if err := c.enviar("recepcion", "POST", "/api/cobros", mapa{"paciente_id": carlos, "cita_id": cobrable, "lineas": []mapa{{"procedimiento_id": d.proc[3], "cantidad": 1}, {"procedimiento_id": d.proc[1], "cantidad": 1}}}, &c1); err != nil {
			return "", err
		}
		if err := c.enviar("recepcion", "POST", fmt.Sprintf("/api/cobros/%d/pagos", c1.ID), mapa{"metodo": "efectivo", "monto_centavos": 3000}, nil); err != nil {
			return "", err
		}
		if err := c.enviar("recepcion", "POST", "/api/cobros", mapa{"paciente_id": ana, "descuento_bp": 1000, "lineas": []mapa{{"procedimiento_id": d.proc[2], "cantidad": 1}}}, &c2); err != nil {
			return "", err
		}
		if err := c.enviar("recepcion", "POST", fmt.Sprintf("/api/cobros/%d/pagos", c2.ID), mapa{"metodo": "tarjeta", "monto_centavos": 3600, "referencia": "AUT-DEMO-0001"}, nil); err != nil {
			return "", err
		}
		if err := c.enviar("recepcion", "POST", "/api/cobros", mapa{"paciente_id": sofia, "lineas": []mapa{{"procedimiento_id": d.proc[1], "cantidad": 1}}}, &c3); err != nil {
			return "", err
		}
		if err := c.enviar("recepcion", "POST", fmt.Sprintf("/api/cobros/%d/anular", c3.ID), mapa{"motivo": "Cobro registrado dos veces por error de captura"}, nil); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s con pago parcial, %s pagado completo (con 10 %% de descuento) y %s anulado", c1.Numero, c2.Numero, c3.Numero), nil
	})

	fmt.Println("\nFase 2: portal del paciente")
	paso("códigos del portal y una cita confirmada por el paciente", func() (string, error) {
		for _, doc := range []string{"0900000001", "0900000003"} {
			if err := c.enviar("", "POST", "/api/portal/codigo", mapa{"documento": doc}, nil); err != nil {
				return "", err
			}
		}
		if err := c.enviar("", "POST", "/api/portal/codigo", mapa{"documento": "0900000002"}, nil); err != nil {
			return "", err
		}
		codigo, err := codigoDe(c, carlos, "otp_portal")
		if err != nil {
			return "", err
		}
		var sesion struct {
			Token string `json:"token"`
		}
		if err := c.enviar("", "POST", "/api/portal/ingresar", mapa{"documento": "0900000002", "codigo": codigo}, &sesion); err != nil {
			return "", err
		}
		c.tokens["paciente"] = sesion.Token
		var citas []struct {
			ID     int    `json:"id"`
			Estado string `json:"estado"`
			Inicio string `json:"inicio"`
			Motivo string `json:"motivo"`
		}
		if err := c.enviar("paciente", "GET", "/api/portal/citas", nil, &citas); err != nil {
			return "", err
		}
		for _, k := range citas {
			inicio, _ := time.Parse(time.RFC3339, k.Inicio)
			if k.Estado == "programado" && inicio.After(time.Now()) && k.Motivo == "Restauración aprobada en el presupuesto" {
				return fmt.Sprintf("códigos enviados a 3 pacientes; Carlos ingresó al portal y confirmó la cita %d", k.ID), c.enviar("paciente", "POST", fmt.Sprintf("/api/portal/citas/%d/confirmar", k.ID), nil, nil)
			}
		}
		return "códigos enviados a 3 pacientes; Carlos ingresó al portal (no había una cita de la demo por confirmar)", nil
	})
}

// examen guarda un examen periodontal de cuatro dientes (11, 16, 26 y 36). nivel desplaza la gravedad de los molares.
func (d *demo) examen(c *cliente, paciente int, fecha string, prof, margen [6]int, sangra [6]bool, nivel int) error {
	mas := func(valores [6]int, extra int) [6]int {
		for i := range valores {
			valores[i] += extra
		}
		return valores
	}
	dientes := []mapa{
		{"pieza": 11, "movilidad": 0, "furca": 0, "sitios": sitios(mas(prof, -1), margen, [6]bool{})},
		{"pieza": 16, "movilidad": 1, "furca": min(nivel, 2), "sitios": sitios(mas(prof, 1), margen, sangra)},
		{"pieza": 26, "movilidad": 0, "furca": 0, "sitios": sitios(prof, margen, sangra)},
		{"pieza": 36, "movilidad": nivel / 3, "furca": nivel / 2, "sitios": sitios(mas(prof, 1), mas(margen, 1), sangra)},
	}
	return c.enviar("odontologo", "POST", ruta(paciente, "/periodontogramas"), mapa{"fecha": fecha, "notas": "Examen de demostración", "dientes": dientes}, nil)
}
