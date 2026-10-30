# Fase 4: UAT, auditoría de seguridad y salida a producción

Sprint 12. Según el documento: pruebas de aceptación de usuario (UAT), auditoría de seguridad y salida a producción.

## Requisitos que cubre

- **RNF-01 (Tiempo de respuesta):** carga del expediente y renderizado del odontograma en menos de 1,5 segundos con conexión de al menos 15 Mbps.
- **RNF-02 (Disponibilidad):** uptime mínimo del 99,9 % en horario operativo clínico (07:00 a 21:00 hora local).
- **RNF-03 (Concurrencia):** 150 usuarios concurrentes sin degradación de latencia.
- **RNF-04 (Cifrado):** TLS 1.3 con certificados vigentes en tránsito; AES-256 en reposo para bases de datos y almacenamiento de archivos.
- **RNF-05 (Autenticación y autorización):** RBAC con menor privilegio (hecho en la etapa 0); MFA obligatoria para acceso a expedientes clínicos e información financiera; cierre de sesión tras 15 minutos de inactividad.
- **RNF-07 (Responsive):** escritorio 1920 × 1080, laptop 1366 × 768 y tablets de 10,5 pulgadas o más.
- **RNF-08 (Navegadores):** las dos últimas versiones estables de Chrome, Firefox, Edge y Safari.
- **8.2 Criterios de aceptación técnica:** cobertura mínima del 80 % de pruebas unitarias y de integración en los endpoints críticos del backend; escaneo DAST y SAST sin vulnerabilidades de severidad alta o crítica (OWASP Top 10); documentación de la API en OpenAPI o Swagger v3.0.
- Respaldos (rol Administrador TI) y salida a producción.

## Punto de partida

Fases 1 a 3 cerradas y desplegadas; auditoría inmutable y pruebas de RN-01, RN-02 y RN-03 ya en la integración continua.

## Issues

### F4-1 MFA y cifrado

**Cubre:** RNF-04, RNF-05 (MFA y cierre de sesión).

**Alcance**
- MFA con TOTP (RFC 6238): alta con código QR, códigos de recuperación de un solo uso, secretos guardados cifrados. El JWT lleva una marca de segundo factor verificado y un middleware `RequireMFA` protege los endpoints de expedientes clínicos e información financiera.
- Cierre de sesión por inactividad de 15 minutos: ya existe en el frontend; se verifica y se refuerza en el servidor (última actividad en Redis) para que no dependa solo del navegador.
- Cifrado en tránsito: verificar TLS 1.3 y certificados vigentes en los dominios de producción con una herramienta de prueba y dejar el resultado documentado. La versión de TLS la determinan Vercel y Render.
- Cifrado en reposo: documentar el cifrado en disco del proveedor de base de datos y de archivos, y añadir cifrado a nivel de aplicación con AES-256-GCM para los datos más sensibles (anamnesis, firmas, archivos), con la clave en una variable de entorno y un procedimiento de rotación. Limitación honesta: PostgreSQL gestionado no ofrece cifrado transparente propio; se apoya en el del proveedor.

**Criterios de aceptación**
- Un usuario sin segundo factor no accede a expedientes ni a información financiera (`403`), y una prueba lo cubre por rol.
- Un código TOTP reutilizado o fuera de ventana se rechaza; un código de recuperación se usa una vez.
- Tras 15 minutos sin actividad, el servidor rechaza el token aunque el navegador siga abierto.
- Un volcado de la base no muestra en claro los campos cifrados.

### F4-2 Cobertura de pruebas, integración continua y OpenAPI

**Cubre:** 8.2 (cobertura mínima del 80 % y documentación OpenAPI).

**Alcance**
- Definir y documentar la lista de endpoints críticos: login y MFA, pacientes y anamnesis, agenda y no solapamiento, evolución y firma, procedimientos realizados, cierre ciego, facturación y pagos, descargo de inventario y comisiones.
- Subir su cobertura (unitaria y de integración) hasta el 80 % o más y fijar un umbral en la integración continua (GitHub Actions) para que no baje.
- Flujo de integración continua: `go vet`, `go test` con cobertura, `npm run build` y `npm run lint` en cada pull request.
- Especificación OpenAPI v3.0 de toda la API, validada en la integración continua y servida como documentación en `/api/docs`.

**Criterios de aceptación**
- El reporte de cobertura de los endpoints críticos muestra 80 % o más y la integración continua falla si baja.
- La especificación OpenAPI valida y cubre todas las rutas existentes.
- Los pull requests ejecutan el flujo completo.

### F4-3 Seguridad: SAST y DAST

**Cubre:** 8.2 (sin vulnerabilidades altas o críticas, OWASP Top 10).

**Alcance**
- SAST: `gosec` y `govulncheck` en el backend, `npm audit` y reglas de seguridad del linter en el frontend, y un escaneo de secretos en el repositorio.
- DAST: OWASP ZAP contra un entorno de pruebas desplegado, con autenticación por rol.
- Revisión del OWASP Top 10: control de acceso (matriz de pruebas por rol, `401` y `403`), cifrado, inyección (consultas parametrizadas), configuración (cabeceras de seguridad, CORS), autenticación (límite de intentos de login), componentes vulnerables y registro de eventos.
- Informe con los hallazgos, su severidad y su corrección.

**Criterios de aceptación**
- Los escaneos SAST y DAST no reportan hallazgos de severidad alta o crítica; los de menor severidad quedan listados con decisión.
- La matriz de permisos por rol está cubierta por pruebas automáticas.
- El login limita los intentos repetidos.

### F4-4 Rendimiento, disponibilidad y compatibilidad

**Cubre:** RNF-01, RNF-02, RNF-03, RNF-07, RNF-08.

**Alcance**
- RNF-01: medir la carga del expediente y el odontograma con red limitada a 15 Mbps y optimizar hasta bajar de 1,5 s.
- RNF-03: prueba de carga con 150 usuarios concurrentes (k6) sobre los flujos principales, con umbrales de latencia fijados y resultados guardados.
- RNF-02: monitoreo de disponibilidad en el horario de 07:00 a 21:00 con un servicio de uptime sobre `GET /api/health`. Se documenta qué plan de hosting cumple el 99,9 %: Render gratuito duerme y Neon gratuito tiene arranques en frío (interpretación 11 del [índice](README.md)).
- RNF-07: verificar el diseño en 1920 × 1080, 1366 × 768 y una tablet de 10,5 pulgadas en vertical y horizontal.
- RNF-08: pruebas automáticas en Chrome, Firefox, Edge y WebKit (referencia de Safari, que no se puede probar directamente en Windows).

**Criterios de aceptación**
- Medición documentada del expediente y el odontograma por debajo de 1,5 s.
- La prueba de carga con 150 usuarios no degrada la latencia por encima del umbral acordado ni produce errores.
- Las pantallas principales se ven y se usan en las tres resoluciones y en los cuatro navegadores.

### F4-5 UAT y salida a producción

**Cubre:** apartado 8.1 (Fase 4), respaldos del rol Administrador TI.

**Alcance**
- Guion de UAT por cada uno de los cinco roles que recorre todos los RF y RN con evidencia (captura o resultado) y registro de hallazgos; correcciones de lo que salga.
- Entorno de producción: secretos propios, CORS y dominios, migraciones `0xx` aplicadas en orden en Neon, sin datos de prueba (`9xx`).
- Respaldos de la base y de los archivos con una restauración probada.
- Runbook: despliegue, reversión, monitoreo y qué hacer ante una caída.
- Manual de usuario breve por rol y documentación final (README, arquitectura). Versión `v1.0.0` etiquetada.

**Criterios de aceptación**
- El guion de UAT se ejecutó completo en producción y no quedan hallazgos abiertos de severidad alta.
- Una restauración de respaldo se probó y su resultado está documentado.
- El sistema responde en el dominio público con TLS y sin datos de prueba.

## Cierre de la Fase 4 y del proyecto

- Los cinco issues integrados y todos los criterios de aceptación de la sección 8.2 del documento cumplidos con evidencia.
- Cada requisito del [índice de trazabilidad](README.md) con su issue cerrado.
- Entrega: repositorio, URL pública, documentación y especificación OpenAPI.
