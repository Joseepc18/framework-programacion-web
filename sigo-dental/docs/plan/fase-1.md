# Fase 1: Core administrativo y turnos

Sprints 1 a 4. Según el documento: pacientes, agendamiento de citas, catálogo médico y facturación básica.

## Requisitos que cubre

- **RF-01 (Registro unificado):** alta, consulta y modificación de la ficha única del paciente: datos demográficos, contacto de emergencia, tutor legal si es menor de edad y seguro dental o convenio.
- **RF-02 (Anamnesis y alertas médicas):** registro obligatorio del historial médico general (enfermedades sistémicas, alergias a fármacos, antecedentes quirúrgicos, coagulopatías). Si hay una condición crítica (por ejemplo alergia a penicilina o anticoagulados), un badge de advertencia de alta visibilidad en la cabecera del expediente.
- **RF-09 (Agendamiento multirrecinto):** calendario por día y semana, filtrable por consultorio o sillón y por profesional.
- **RF-10 (Estados de cita):** Programado, Confirmado, En Sala de Espera, En Atención, Atendido, Cancelado y No Asistió.
- **RF-11 (Notificaciones automáticas):** confirmación y recordatorio 48 h y 24 h antes del turno por WhatsApp, SMS o correo.
- **RF-12 (parte básica):** cobro y comprobante. El comprobante fiscal electrónico y los pagos mixtos van en la Fase 3.
- **RN-03 (Sobreposición de sillones):** no se pueden agendar dos citas simultáneas en el mismo sillón físico, aunque los odontólogos sean distintos.
- Catálogo médico (mencionado por el documento en la fase, sin número de requisito).

Roles involucrados: recepcionista o cajero (registro de pacientes, agendamiento, cobro), odontólogo y asistente (lectura del paciente y de la agenda), administrador TI (parametrización de catálogos) y auditor (lectura de cobros).

## Punto de partida

La etapa 0: API con login y roles, SPA con layout, login y menú por rol. No existe ninguna tabla de negocio.

## Issues

### F1-1 Catálogo médico

**Cubre:** catálogo médico (interpretación 1 del [índice](README.md)), parametrización de catálogos del rol Administrador TI.

**Alcance**
- Profesionales: odontólogos vinculados a un usuario, con especialidad y número de registro profesional.
- Consultorios y sillones: cada sillón físico pertenece a un consultorio y puede estar activo o inactivo.
- Procedimientos: código, nombre, especialidad, precio base, duración estimada en minutos y marca de "requiere consentimiento" (la usa F2-5).
- Convenios y seguros dentales.
- API y pantallas de alta, consulta, modificación y desactivación. Nada se borra: se desactiva, porque otros registros lo referencian.
- Datos de prueba para desarrollo en un script `9xx`.

**Permisos:** el administrador TI gestiona los catálogos; el resto de los roles solo los consulta.

**Criterios de aceptación**
- Un usuario que no es administrador TI recibe `403` al intentar crear o modificar un catálogo.
- No se puede crear dos veces el mismo código de procedimiento ni dos sillones con el mismo nombre en un consultorio.
- Un elemento desactivado deja de ofrecerse en los selectores pero sigue visible en los registros que ya lo usan.
- Pruebas de la API (permisos y duplicados) y migración numerada aplicada en local.

### F1-2 Pacientes, anamnesis y alertas médicas

**Cubre:** RF-01, RF-02.

**Alcance**
- Ficha única con número de historia clínica autogenerado y único, documento de identidad, nombres, fecha de nacimiento, sexo, teléfono, correo y dirección.
- Contacto de emergencia; tutor legal, obligatorio cuando la edad calculada es menor de 18 años; seguro dental o convenio (del catálogo de F1-1).
- Anamnesis obligatoria al registrar: enfermedades sistémicas, alergias a fármacos, antecedentes quirúrgicos y coagulopatías. Cada apartado se completa o se marca expresamente como "ninguna"; no puede quedar vacío.
- Lista configurable de condiciones críticas (alergia a penicilina, anticoagulados, coagulopatías, entre otras). Si el paciente tiene alguna, se muestra un badge de alta visibilidad en la cabecera del expediente y en el detalle de sus citas.
- Búsqueda por nombre, documento o número de historia. Lista, alta, edición y ficha con cabecera del expediente (la cabecera se reutiliza en la Fase 2).
- Los pacientes no se borran: se desactivan.

**Permisos:** la recepcionista da de alta y modifica datos demográficos; el odontólogo y el asistente leen la ficha y editan la anamnesis. El auditor no accede a datos clínicos.

**Criterios de aceptación**
- No se puede guardar un paciente sin anamnesis completa ni un menor de edad sin tutor legal.
- Un documento de identidad repetido se rechaza con un mensaje claro.
- Un paciente con una condición crítica muestra el badge en todas las pantallas de su expediente, con texto y no solo color.
- Los permisos por rol están cubiertos por pruebas.
- La pantalla funciona en laptop y tablet (RNF-07).

### F1-3 Agenda, estados de cita y no solapamiento

**Cubre:** RF-09, RF-10, RN-03.

**Alcance**
- Citas: paciente, profesional, sillón, inicio, fin, procedimiento o motivo, estado y notas. Las fechas se guardan en UTC y se muestran en hora local; el horario de la agenda es configurable (por defecto de 07:00 a 21:00, el horario operativo del documento).
- Calendario por día y por semana, filtrable por consultorio o sillón y por profesional. Crear desde un hueco del calendario, reprogramar y cancelar con motivo.
- Siete estados: Programado, Confirmado, En Sala de Espera, En Atención, Atendido, Cancelado y No Asistió. Transiciones permitidas: de Programado a Confirmado, En Sala de Espera, Cancelado o No Asistió; de Confirmado a En Sala de Espera, Cancelado o No Asistió; de En Sala de Espera a En Atención o Cancelado; de En Atención a Atendido. Atendido, Cancelado y No Asistió son finales. Cada cambio queda en un historial con quién y cuándo.
- RN-03 se aplica en la base de datos con una restricción de exclusión sobre el rango de tiempo por sillón, que ignora las citas Canceladas y No Asistió. Citas consecutivas (una termina cuando empieza la otra) son válidas.
- Estilos de los siete estados en los tokens de Aurora. `docs/diseno.md` define tres; se añaden los cuatro que faltan, con texto además de color.

**Permisos:** la recepcionista agenda y cambia cualquier estado; el odontólogo y el asistente cambian a En Atención y Atendido y consultan la agenda.

**Criterios de aceptación**
- Dos solicitudes simultáneas para el mismo sillón y horario: una se crea y la otra recibe `409`, incluso si los odontólogos son distintos.
- Una cita cancelada libera el sillón; una transición no permitida devuelve `409` con el motivo.
- Los filtros por sillón y por profesional devuelven solo lo pedido.
- Pruebas de solapamiento (mismo sillón, distinto odontólogo, citas contiguas, cita cancelada) y de transiciones.
- Decisión pendiente: el documento no prohíbe que un mismo profesional tenga dos citas a la vez en sillones distintos. Se propone avisar y no bloquear.

### F1-4 Notificaciones automáticas

**Cubre:** RF-11 y el objetivo de reducir el ausentismo (apartado 2.2).

**Alcance**
- Interfaz `Notificador` con los canales WhatsApp, SMS y correo, y un adaptador simulado que registra cada mensaje en una tabla `notificaciones` con su estado (enviada o fallida). Los adaptadores reales (Meta Cloud API, Twilio) quedan para cuando haya credenciales.
- Mensaje de confirmación al agendar y recordatorios 48 h y 24 h antes del turno. Un planificador revisa las citas pendientes y envía cada mensaje una sola vez (clave única por cita y tipo).
- Si la cita se cancela, no se envían más recordatorios; si se reprograma, se regeneran.
- Enlace de confirmación con token firmado que pasa la cita a Confirmado (propuesta para atacar el ausentismo).
- Pantalla con la bitácora de notificaciones por cita.

**Permisos:** la recepcionista ve la bitácora; el canal preferido sale de la ficha del paciente.

**Criterios de aceptación**
- Una cita crea exactamente una confirmación y dos recordatorios, aunque el planificador se ejecute varias veces.
- Una cita cancelada no genera recordatorios; una reprogramada recalcula los suyos.
- Cambiar del adaptador simulado a uno real es solo configuración.
- Nota: Render gratuito duerme tras 15 minutos sin tráfico y puede retrasar recordatorios; queda documentado (ver RNF-02).

### F1-5 Facturación básica

**Cubre:** RF-12 (parte básica).

**Alcance**
- Cobro con líneas de procedimiento (cantidad y precio del catálogo), descuento y total. Se puede generar desde una cita Atendida.
- Pagos en efectivo, tarjeta o transferencia con referencia, incluyendo pago parcial. Estados: pendiente, pagado parcial y pagado.
- Comprobante interno con numeración secuencial, vista imprimible.
- Anulación como registro nuevo; nada se borra.
- Dinero como entero en centavos o `numeric`, nunca como decimal flotante. Moneda configurable (decisión de país pendiente).

**Permisos:** la recepcionista o cajero cobra; el auditor consulta.

**Criterios de aceptación**
- El total se calcula en el servidor; el cliente no lo decide.
- Un cobro pagado no se puede modificar; solo anular.
- Los números de comprobante no se repiten ni se saltan bajo solicitudes simultáneas.
- Pruebas de redondeo, pago parcial, anulación y permisos.

## Cierre de la Fase 1

- Los cinco issues integrados y desplegados en Vercel, Render y Neon con las migraciones aplicadas.
- Recorrido de aceptación con una cuenta de prueba por rol: registrar un paciente con anamnesis, agendar su cita sin choque de sillón, recibir la notificación simulada, atenderla y cobrarla.
- `go test ./...`, `go vet ./...`, `npm run build` y `npm run lint` sin errores.
- README, `docs/architecture.md` y `CLAUDE.md` actualizados con lo nuevo.
