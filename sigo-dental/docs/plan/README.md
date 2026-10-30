# Plan de implementación

Plan completo derivado del Documento de Requerimientos de Software v1.0 (29/10/2026). Hay un archivo por fase del apartado 8.1 del documento; cada uno detalla lo que debe cumplirse y se divide en issues del tamaño de una sesión de trabajo.

- [Fase 1: Core administrativo y turnos](fase-1.md) (sprints 1 a 4)
- [Fase 2: Core clínico](fase-2.md) (sprints 5 a 8)
- [Fase 3: Finanzas avanzadas, inventarios y reportes](fase-3.md) (sprints 9 a 11)
- [Fase 4: UAT, auditoría de seguridad y salida a producción](fase-4.md) (sprint 12)

## Cómo se usa

- Cada ítem `F<fase>-<n>` pasa a ser un issue de GitHub con ese ID en el título. El issue cubre solo lo que su ítem describe.
- Una fase termina cuando todos sus issues están integrados, desplegados y cumplen los criterios de cierre de su archivo.
- Si durante la implementación aparece algo que el plan no cubre, se corrige este plan en el mismo cambio, para que siga siendo la referencia.

## Base ya construida (etapa 0)

Antes de la Fase 1 existe la base técnica, que cubre parcialmente RNF-05 y RNF-07:

- API en Go con `GET /api/health`, usuarios y roles (cinco roles del documento), login con JWT y control de acceso por roles (middleware `RequireRoles`).
- SPA en React con layout responsive, dirección visual Aurora, login, rutas protegidas por rol, menú por rol y cierre de sesión por inactividad.

Pendiente de la etapa 0:

- **E0-1 Despliegue inicial.** Frontend en Vercel (raíz `sigo-dental/frontend`), API en Render con Docker (raíz `sigo-dental/backend`) y Redis, base en Neon con las migraciones `0xx` aplicadas a mano. Variables de entorno, CORS hacia el dominio de Vercel y `GET /api/health` respondiendo en producción. Se documenta en `docs/despliegue.md`. Cada fase se despliega al cerrarse.

## Trazabilidad

| Requisito | Issue |
|---|---|
| RF-01 Registro unificado | F1-2 |
| RF-02 Anamnesis y alertas médicas | F1-2 |
| RF-03 Odontograma | F2-2 |
| RF-04 Periodontograma | F2-3 |
| RF-05 Evolución y notas SOEP | F2-4 |
| RF-06 Consentimiento informado digital | F2-5 |
| RF-07 Presupuestos multinivel | F2-7 |
| RF-08 Aprobación y transición a citas | F2-7 |
| RF-09 Agendamiento multirrecinto | F1-3 |
| RF-10 Estados de cita | F1-3 |
| RF-11 Notificaciones automáticas | F1-4 |
| RF-12 Punto de venta y facturación electrónica | F1-5 (básico) y F3-1 (completo) |
| RF-13 Liquidación de honorarios | F3-4 |
| RF-14 Descargo automático de insumos | F3-3 |
| RNF-01 Tiempo de respuesta | F2-2 (diseño) y F4-4 (medición) |
| RNF-02 Disponibilidad | F4-4 |
| RNF-03 Concurrencia | F4-4 |
| RNF-04 Cifrado | F4-1 |
| RNF-05 Autenticación y autorización | etapa 0 (RBAC, cierre por inactividad) y F4-1 (MFA) |
| RNF-06 Trazabilidad y auditoría | F2-1 |
| RNF-07 Diseño responsive | etapa 0 y F4-4 (verificación) |
| RNF-08 Soporte de navegadores | F4-4 |
| RN-01 Inmutabilidad del historial clínico | F2-4 |
| RN-02 Cierre de caja ciego | F3-2 |
| RN-03 Sobreposición de sillones | F1-3 |
| 8.2 Cobertura mínima del 80 % | F4-2 |
| 8.2 DAST/SAST sin severidad alta o crítica | F4-3 |
| 8.2 Documentación OpenAPI v3.0 | F4-2 |

Elementos que el documento menciona sin numerarlos como requisito:

| Elemento | Issue |
|---|---|
| Catálogo médico (Fase 1) | F1-1 |
| Recetas (Fase 2) | F2-6 |
| Firma digital (Fase 2) | F2-4 |
| Registro de signos vitales (rol Asistente) | F2-1 |
| Almacenamiento de archivos: radiografías, fotos y PDF (apartado 7) | F2-1 |
| Módulo de telemetría y auditoría (apartado 2.3) | F2-1 |
| Portal PWA del paciente (apartado 2.3) | F2-8 |
| Reportes BI (Fase 3) | F3-5 |
| Respaldos y parametrización de catálogos (rol Administrador TI) | F4-5 y F1-1 |
| Inventario por consultorio y por profesional (apartado 2.2) | F3-3 |
| Reducción del ausentismo (apartado 2.2) | F1-4 |

## Interpretaciones y decisiones pendientes

El documento deja huecos o se contradice en algunos puntos. Esto es lo que se asumió; cambiar cualquiera de estos puntos es corregir el plan, no el código.

1. **Catálogo médico.** El documento lo nombra en la Fase 1 sin definirlo. Se entiende como profesionales, consultorios y sillones, procedimientos con precio y convenios o seguros.
2. **RF-07 y RF-08 sin fase.** El documento no asigna los presupuestos a ninguna fase. Se ubican en la Fase 2 porque dependen del odontograma y de la agenda.
3. **Facturación en dos tiempos.** La Fase 1 pide "facturación básica" y la Fase 3 "finanzas avanzadas". RF-12 se parte así: cobro con comprobante interno en la Fase 1 y comprobante electrónico, pagos mixtos y pasarela en la Fase 3.
4. **Integraciones externas simuladas.** Facturación fiscal, mensajería (WhatsApp, SMS y correo) y pasarela de pagos se implementan como interfaces con un adaptador simulado. Los servicios reales exigen credenciales, certificados o costos; cambiarlos después no debe tocar el resto del código.
5. **Portal PWA del paciente.** El apartado 2.3 dice que se aborda en la Fase 2 pero no define requisitos. F2-8 propone un mínimo y exige un rol nuevo (paciente), que no está en la tabla de roles del documento. Por confirmar.
6. **Firma digital.** El documento no fija el nivel. Se asume firma electrónica con huella criptográfica del contenido, identidad del firmante y fecha, no certificado de firma electrónica avanzada. Por confirmar.
7. **Restauración en azul.** RF-03 pide restauraciones en azul y `docs/diseno.md` las define en turquesa. Hay que alinear uno de los dos.
8. **Estados de cita sin estilo.** `docs/diseno.md` solo define el estilo de tres estados; RF-10 exige siete. F1-3 completa los cuatro que faltan.
9. **Auditoría adelantada.** El log de auditoría (RNF-06) se construye en la Fase 2, junto con el expediente, y no al final, porque debe registrar cada acceso desde que existen datos clínicos. La Fase 4 lo verifica.
10. **MFA al final.** El documento pone la auditoría de seguridad en la Fase 4 y por eso MFA y cifrado van ahí. Puede adelantarse si hace falta.
11. **Disponibilidad del 99,9 %.** Render gratuito duerme tras 15 minutos sin tráfico y Neon gratuito tiene arranques en frío; con eso no se puede garantizar RNF-02. Se mide y se documenta qué plan lo cumpliría.
12. **País y moneda.** Sin decidir (facturación fiscal, ley de protección de datos). Moneda configurable.
13. **Errores del documento.** La fórmula de RNF-01 aparece corrupta; se lee como "menos de 1,5 segundos con conexión de al menos 15 Mbps". Las firmas de aprobación del apartado 1.2 están en blanco.
