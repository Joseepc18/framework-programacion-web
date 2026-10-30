# Fase 2: Core clínico

Sprints 5 a 8. Según el documento: odontograma interactivo, periodontograma, recetas, consentimientos informados y firma digital. El apartado 2.3 añade el portal PWA del paciente a esta fase.

## Requisitos que cubre

- **RF-03 (Odontograma gráfico interactivo):** dentición permanente (32 dientes) y temporal (20); nomenclatura FDI y Universal; marcado por superficies (mesial, distal, oclusal o incisal, vestibular, lingual o palatino); catálogo visual: caries (rojo), restauraciones existentes (azul), endodoncia realizada o requerida, implantes, prótesis fijas o removibles, exodoncias y sellantes.
- **RF-04 (Periodontograma digital):** profundidad al sondaje (6 puntos por diente), margen gingival, cálculo del nivel de inserción clínica (NIC), furca, movilidad y sangrado al sondaje.
- **RF-05 (Evolución y notas SOEP):** bitácora cronológica inmutable (Subjetivo, Objetivo, Evaluación, Plan). Cada nota se bloquea contra edición pasado un plazo configurable (24 horas por defecto) y requiere firma digital del profesional.
- **RF-06 (Consentimiento informado digital):** formatos legales por procedimiento, con firma biométrica en tablet o pantalla táctil, o validación con código OTP enviado al paciente.
- **RF-07 (Presupuestos multinivel):** tratamientos por fases y sesiones, con honorarios, costo de laboratorio y descuentos autorizados.
- **RF-08 (Aprobación y transición a citas):** aprobación total o parcial del presupuesto; los ítems autorizados pasan a ser tareas agendables.
- **RNF-06 (Trazabilidad y auditoría):** log inmutable de cada acción CRUD sobre historias clínicas: quién consultó, modificó o exportó qué registro, con timestamp ISO-8601 y dirección IP.
- **RNF-01 (en lo que toca a esta fase):** el expediente y el odontograma deben cargar en menos de 1,5 segundos.
- **RN-01 (Inmutabilidad del historial clínico):** ningún procedimiento "Realizado" se borra; las correcciones se hacen con una nota de adenda firmada por el profesional responsable.
- Recetas, firma digital, registro de signos vitales (rol Asistente), almacenamiento de archivos (radiografías, fotos y PDF) y portal PWA del paciente: mencionados por el documento sin número de requisito.

Roles involucrados: odontólogo (registro clínico, odontograma, plan de tratamiento, recetas, evolución), asistente (signos vitales, preparación de insumos, notas clínicas), recepcionista (presupuestos y agenda), administrador TI y auditor (lectura de la auditoría).

## Punto de partida

Fase 1 cerrada: pacientes con anamnesis y cabecera de expediente, catálogo, agenda y cobros.

## Issues

### F2-1 Expediente clínico, auditoría inmutable y archivos

**Cubre:** RNF-06, registro de signos vitales, almacenamiento de archivos, módulo de telemetría y auditoría (apartado 2.3).

**Alcance**
- Vista de expediente con pestañas para los módulos clínicos; las pestañas se habilitan a medida que llegan los siguientes issues. La cabecera con el badge de alertas ya existe (F1-2).
- Signos vitales por visita (presión arterial, frecuencia cardiaca y respiratoria, temperatura, saturación, peso y talla), registrados por el asistente o el odontólogo.
- Archivos del paciente (radiografías, fotos, PDF) detrás de una interfaz `AlmacenArchivos` con un adaptador de disco local para desarrollo y otro compatible con S3 para producción. Metadatos en la base, descarga solo con sesión válida y permiso, límites de tamaño y de tipo. Sin visor DICOM (fuera de alcance del documento).
- Tabla `auditoria_accesos` de solo inserción: el usuario de la aplicación no tiene permiso de `UPDATE` ni `DELETE`, y un disparador lo impide. Cada consulta, creación, modificación o exportación de un registro clínico guarda quién, qué acción, qué entidad y paciente, timestamp ISO-8601 en UTC y dirección IP de origen (respetando el proxy de Render). Se aplica también a los endpoints de pacientes y anamnesis de la Fase 1.
- Pantalla de consulta de la auditoría con filtros por usuario, paciente, acción y fechas.

**Permisos:** administrador TI y auditor leen la auditoría; nadie la modifica.

**Criterios de aceptación**
- Intentar actualizar o borrar una fila de auditoría falla, incluso desde la propia API.
- Toda lectura de una ficha o expediente deja una fila con la IP y el usuario correctos.
- Un archivo no se descarga sin sesión ni permiso.
- Decisión pendiente: proveedor S3 compatible para producción (el disco de Render gratuito es efímero).

### F2-2 Odontograma interactivo

**Cubre:** RF-03, RNF-01 (renderizado).

**Alcance**
- Odontograma en SVG con 32 dientes permanentes y 20 temporales, selector permanente, temporal o mixta (propuesta para pacientes en recambio).
- Alternar entre numeración FDI y Universal; internamente se guarda siempre FDI.
- Cinco superficies por diente: mesial, distal, oclusal o incisal, vestibular, lingual o palatino.
- Catálogo de hallazgos: caries, restauración existente, endodoncia (realizada o requerida), implante, prótesis fija, prótesis removible, exodoncia, sellante y diente ausente. Colores de `docs/diseno.md`; el color nunca es el único indicador: cada estado lleva leyenda, texto o forma.
- Cada hallazgo guarda pieza, superficie, tipo, estado, quién lo registró y cuándo. El odontograma conserva historial por visita; no se sobrescribe.
- Un hallazgo "requerido" se puede corregir o retirar; uno "realizado" no (RN-01, ver F2-4).
- Controles táctiles de 44 px como mínimo, para uso en tablet junto al sillón.

**Criterios de aceptación**
- El expediente con odontograma carga en menos de 1,5 s con conexión de al menos 15 Mbps (se mide con el throttling de red del navegador).
- Los 52 dientes y las cinco superficies se pueden marcar con el dedo en una tablet de 10,5 pulgadas.
- Alternar FDI y Universal no cambia los datos guardados.
- Decisión pendiente: el PDF pide restauraciones en azul y `docs/diseno.md` las define turquesa; se alinea antes de implementar.
- Pruebas de la API (permisos, historial) y del componente (marcado, numeración).

### F2-3 Periodontograma digital

**Cubre:** RF-04.

**Alcance**
- Por cada diente presente, seis puntos de sondaje (tres vestibulares y tres linguales): profundidad al sondaje en mm, margen gingival en mm y sangrado al sondaje.
- Por diente: furca (grado I a III) y movilidad (0 a 3).
- Cálculo del NIC en el servidor a partir de la profundidad y el margen gingival, con la convención de signo fijada y documentada en el issue, y verificado con pruebas de valores conocidos.
- Entrada pensada para el sillón: teclado numérico y avance automático entre puntos. Vista en tabla y gráfica por arcada.
- Exámenes con fecha, para comparar la evolución. Los dientes ausentes del odontograma (F2-2) no se ofrecen.

**Criterios de aceptación**
- Valores fuera de rango (por ejemplo, profundidad negativa o mayor que 15 mm) se rechazan.
- El NIC coincide con los casos de prueba calculados a mano.
- Se puede capturar un examen completo de 32 dientes sin usar el ratón.

### F2-4 Evolución SOEP, adendas, firma digital e inmutabilidad

**Cubre:** RF-05, RN-01, firma digital.

**Alcance**
- Nota SOEP con los cuatro campos, ligada al paciente y a la cita o procedimiento. Bitácora cronológica por paciente.
- Parámetro `bloqueo_nota_horas` (24 por defecto). Pasado el plazo, el servidor rechaza cualquier edición.
- Firma digital del profesional: reautenticación con su contraseña, huella SHA-256 del contenido canónico, identidad y fecha. La firma bloquea la nota de inmediato; una nota sin firmar se bloquea al vencer el plazo y el autor aún puede firmarla (queda en "pendiente de firma" hasta entonces). Este servicio de firma lo reutilizan F2-5 y F2-6. Nivel de firma: interpretación 6 del [índice](README.md).
- Registro de procedimientos realizados (paciente, procedimiento del catálogo, pieza y superficie, profesional, fecha). Una vez "Realizado" no se puede borrar ni cambiar; no existe endpoint de borrado y un disparador de la base lo impide.
- Adenda: nota nueva, enlazada a la original y firmada por el profesional responsable, es la única forma de corregir.

**Permisos:** el odontólogo firma; el asistente puede redactar notas clínicas pero no firmarlas.

**Criterios de aceptación**
- Editar una nota después del plazo o después de firmarla devuelve error, y se verifica en la prueba con el reloj controlado.
- Intentar borrar o modificar un procedimiento "Realizado" falla a nivel de API y de base de datos.
- Alterar el contenido de una nota firmada hace que su huella deje de coincidir y se detecte.
- Todo acceso y cambio queda en la auditoría de F2-1.

### F2-5 Consentimiento informado digital

**Cubre:** RF-06.

**Alcance**
- Plantillas de consentimiento por procedimiento (los marcados "requiere consentimiento" en el catálogo) con campos que se completan solos: paciente, procedimiento, profesional y fecha.
- Dos formas de firmar: captura de firma manuscrita en pantalla táctil (trazo, marca de tiempo y dispositivo) o código OTP enviado al paciente por el `Notificador` de F1-4 (guardado con hash, vence a los 10 minutos, máximo tres intentos). Si el paciente es menor, firma el tutor legal.
- Documento PDF del consentimiento firmado, guardado en `AlmacenArchivos` e inmutable. Una revocación es un documento nuevo.
- Con el adaptador simulado, el OTP queda visible en la bitácora de notificaciones para poder probar.

**Criterios de aceptación**
- No se puede iniciar un procedimiento marcado "requiere consentimiento" sin un consentimiento firmado vigente (aviso visible, definido en el issue).
- Un OTP vencido, usado o con tres intentos fallidos no valida.
- El PDF firmado no se puede modificar y queda en la auditoría.

### F2-6 Recetas

**Cubre:** recetas (Fase 2) y el objetivo de digitalizar el 100 % de las recetas (apartado 2.2).

**Alcance**
- Receta con medicamentos (nombre, dosis, frecuencia, duración e indicaciones), numerada, ligada al paciente y a la atención.
- Aviso al recetar si el medicamento coincide con una alergia registrada en la anamnesis (propuesta; el documento no lo exige).
- Firma del odontólogo con el servicio de F2-4. Una receta firmada no se edita; se anula con otra.
- Vista imprimible y PDF con el encabezado de la clínica. Historial de recetas por paciente.

**Permisos:** solo el odontólogo emite recetas.

**Criterios de aceptación**
- Una receta firmada no se puede editar; anularla deja registro.
- Recetar un fármaco al que el paciente es alérgico muestra el aviso.
- La numeración no se repite bajo solicitudes simultáneas.

### F2-7 Presupuestos y aprobación a citas

**Cubre:** RF-07, RF-08.

**Alcance**
- Presupuesto por fases y sesiones: cada ítem lleva procedimiento del catálogo, pieza o superficie (opcionalmente ligado a un hallazgo del odontograma), honorarios (editables sobre el precio base), costo de laboratorio y descuento. Los descuentos requieren autorización de un rol habilitado y se registra quién autorizó. Totales calculados en el servidor.
- Estados: borrador, presentado, aprobado total, aprobado parcial y rechazado. Un presupuesto aprobado no se edita; un cambio genera una versión nueva.
- Aprobación total o parcial por ítem, con registro de quién y cuándo (y firma o OTP de F2-5 si se decide exigirla).
- Los ítems autorizados pasan a una lista de tareas agendables del paciente; desde ahí se crea la cita en la agenda (F1-3) con el procedimiento y la duración ya cargados. Al completarse el procedimiento (F2-4), el ítem queda cumplido.
- Presupuesto imprimible en PDF.

**Permisos:** el odontólogo arma el plan; la recepcionista presenta y registra la aprobación.

**Criterios de aceptación**
- Aprobar parcialmente genera tareas solo para los ítems aprobados.
- El total del presupuesto coincide con la suma de ítems, descuentos y costos, verificado con pruebas.
- Un descuento sin autorización válida se rechaza.

### F2-8 Portal PWA del paciente

**Cubre:** apartado 2.3 (portal PWA en Fase 2). Sin requisitos definidos por el documento: este alcance es una propuesta por confirmar (interpretación 5 del [índice](README.md)).

**Alcance mínimo propuesto**
- Rol nuevo `paciente`, con ingreso por documento de identidad y OTP.
- Ver sus citas, confirmarlas o cancelarlas (alimenta RF-10 y RF-11), ver sus recetas y consentimientos firmados en solo lectura.
- Aplicación instalable: manifiesto, service worker y funcionamiento básico sin conexión del cascarón. Optimizada para teléfono.

**Criterios de aceptación**
- Un paciente solo ve y modifica sus propios datos.
- Confirmar una cita desde el portal pasa su estado a Confirmado y queda en el historial.
- Se instala y abre desde la pantalla de inicio de un teléfono.

## Cierre de la Fase 2

- Los ocho issues integrados y desplegados con las migraciones aplicadas.
- Recorrido de aceptación: el odontólogo abre un expediente, carga el odontograma y el periodontograma, firma una nota SOEP, emite una receta, obtiene un consentimiento firmado y aprueba parcialmente un presupuesto que se vuelve cita; la auditoría muestra todo.
- Prueba de RN-01 y del bloqueo de notas pasando en la integración continua.
- Documentación actualizada (README, arquitectura, `CLAUDE.md`).
