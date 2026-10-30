# Fase 3: Finanzas avanzadas, inventarios y reportes

Sprints 9 a 11. Según el documento: descargo de insumos, comisiones médicas y reportes BI.

## Requisitos que cubre

- **RF-12 (Punto de venta y facturación electrónica):** emisión de comprobantes fiscales electrónicos autorizados por la autoridad tributaria local, con pagos mixtos (efectivo, tarjeta de crédito o débito, transferencias, seguros). El apartado 7.1 añade timbrado de facturas y retenciones, y la pasarela de pagos.
- **RN-02 (Cierre de caja ciego):** los recepcionistas no pueden ver el total teórico del sistema antes de ingresar el conteo físico de efectivo en el cierre diario.
- **RF-13 (Liquidación de honorarios):** cálculo configurable de comisiones a odontólogos: porcentaje fijo, monto fijo por procedimiento o porcentaje tras deducir costos de laboratorio y materiales.
- **RF-14 (Descargo automático de insumos):** al completar un procedimiento clínico, se descuenta el stock de insumos según la receta estándar de consumo configurada (resinas, anestesias, guantes).
- Inventario por consultorio y por profesional (objetivo del apartado 2.2) y reportes BI (sin número de requisito).

Roles involucrados: recepcionista o cajero (cobro y cierre), auditor o contador (reportes de corte de caja, liquidación de honorarios e inventario), asistente (preparación de insumos), administrador TI (configuración).

## Punto de partida

Fase 2 cerrada. En particular, los procedimientos "Realizados" (F2-4) son la fuente de los descargos de inventario y de las comisiones, y los cobros básicos de F1-5 son la base de la facturación completa.

## Issues

### F3-1 Facturación electrónica y pagos mixtos

**Cubre:** RF-12 (completo).

**Alcance**
- Un cobro admite varios pagos de distinto tipo: efectivo, tarjeta de crédito o débito, transferencia y seguro (la parte cubierta por el convenio queda como cuenta por cobrar a la aseguradora).
- Interfaz `FacturacionElectronica` (emitir, consultar estado, anular y registrar retenciones) con un adaptador simulado que genera número de autorización y representación del comprobante. El adaptador real exige certificado de firma y credenciales ante la autoridad tributaria; queda fuera hasta tenerlos (interpretación 4 del [índice](README.md)).
- Interfaz `PasarelaPagos` (crear cobro, confirmar, reembolsar) con adaptador simulado para tarjeta, con claves de idempotencia para no cobrar dos veces. Los proveedores reales que cita el documento son Stripe, dLocal o pasarelas bancarias.
- Datos fiscales del cliente (identificación y razón social) en el paciente o en el cobro.
- Un comprobante emitido no se modifica: se anula con una nota de crédito.

**Permisos:** la recepcionista emite y cobra; el auditor consulta.

**Criterios de aceptación**
- La suma de los pagos de un cobro es igual a su total antes de emitir el comprobante.
- Repetir la misma solicitud de pago con la misma clave de idempotencia no genera un segundo cobro.
- Un comprobante emitido no se puede editar ni borrar; la anulación deja registro.
- Pruebas de pagos mixtos, parte cubierta por seguro y anulación.

### F3-2 Caja y cierre ciego

**Cubre:** RN-02.

**Alcance**
- Apertura de caja con fondo inicial, movimientos de efectivo (cobros, ingresos y egresos) y cierre diario.
- En el cierre, la recepcionista ingresa el conteo físico antes de ver el total teórico. Solo después de registrarlo, el sistema revela el total teórico y la diferencia (sobrante o faltante).
- El auditor y el administrador TI sí pueden ver el total teórico en cualquier momento.
- Los cierres son inmutables y generan el reporte de corte de caja.

**Criterios de aceptación**
- Ninguna respuesta de la API (detalle de caja, listados, reportes, exportaciones) incluye el total teórico para un recepcionista con la caja abierta. Hay una prueba que recorre todos los endpoints de caja con ese rol.
- Un cierre ya registrado no se modifica.
- La diferencia se calcula en el servidor y se guarda con el conteo.

### F3-3 Inventario y descargo automático de insumos

**Cubre:** RF-14 y el inventario por consultorio y profesional del apartado 2.2.

**Alcance**
- Catálogo de insumos (resinas, anestesias, guantes, entre otros) con unidad de medida, costo unitario y stock mínimo.
- Stock por consultorio y por profesional. Kardex de movimientos: ingreso, ajuste, transferencia y consumo. Los movimientos no se borran; los errores se corrigen con un movimiento contrario.
- Receta estándar de consumo por procedimiento del catálogo: lista de insumos y cantidades.
- Descargo automático cuando un procedimiento pasa a "Realizado" (F2-4): se descuenta el stock en la misma transacción y una sola vez por procedimiento.
- Alertas de stock mínimo. Política ante stock insuficiente: el procedimiento no se bloquea, el stock queda marcado en negativo y se avisa (decisión pendiente de confirmar).

**Permisos:** el asistente prepara y consume insumos; el auditor consulta inventario; el administrador TI configura recetas de consumo.

**Criterios de aceptación**
- Registrar un procedimiento realizado descuenta exactamente lo que dice su receta de consumo.
- Si el descargo falla, el procedimiento no queda registrado a medias (transacción única).
- Un mismo procedimiento no descuenta stock dos veces aunque se repita la solicitud.
- El costo de insumos consumidos queda disponible para las comisiones (F3-4).

### F3-4 Comisiones y liquidación de honorarios

**Cubre:** RF-13.

**Alcance**
- Configuración por profesional (y por procedimiento si se desea) de una de tres modalidades: porcentaje fijo, monto fijo por procedimiento, o porcentaje tras deducir costos de laboratorio y materiales (el laboratorio viene del presupuesto de F2-7 y los materiales del consumo de F3-3).
- Liquidación por período a partir de los procedimientos realizados y cobrados. Base de cálculo, cobrado o facturado: decisión pendiente, por defecto cobrado.
- Estados de la liquidación: borrador, cerrada y pagada. Una liquidación cerrada no se modifica.
- Reporte de liquidación por profesional con el detalle de cada procedimiento.

**Permisos:** el auditor o contador genera y cierra liquidaciones; el odontólogo ve las suyas.

**Criterios de aceptación**
- Cada modalidad tiene pruebas con valores calculados a mano, incluido el caso con costos deducidos.
- Cambiar la configuración de comisión no altera liquidaciones ya cerradas.
- Un odontólogo no ve las liquidaciones de otro.

### F3-5 Reportes BI

**Cubre:** reportes BI (Fase 3). El documento no enumera los reportes; se proponen los siguientes, que sirven a los objetivos del apartado 2.2.

**Alcance**
- Agenda: citas por estado y tasa de ausentismo (No Asistió) por período, profesional y consultorio.
- Finanzas: ingresos por período, método de pago, profesional y convenio; cuentas por cobrar a aseguradoras; cortes de caja y diferencias.
- Producción: procedimientos realizados por profesional y por tipo.
- Inventario: consumo por consultorio y por profesional, y artículos bajo el stock mínimo.
- Comisiones liquidadas por período.
- Filtros por fechas, consultorio y profesional. Exportación a CSV y PDF. Consultas con vistas SQL y caché en Redis para que sigan rápidas con muchos datos.

**Permisos:** el auditor y el administrador TI ven todo; la recepcionista solo lo operativo de agenda y nunca totales de caja (RN-02).

**Criterios de aceptación**
- Las cifras de cada reporte coinciden con los datos de origen en pruebas con un conjunto conocido.
- La caché se invalida al cambiar los datos de origen.
- Un rol sin permiso recibe `403` en cada reporte financiero.

## Cierre de la Fase 3

- Los cinco issues integrados y desplegados con las migraciones aplicadas.
- Recorrido de aceptación: cobro con pago mixto y comprobante electrónico simulado, cierre de caja ciego, un procedimiento realizado que descuenta stock, una liquidación de honorarios cerrada y los reportes cuadrando con esos datos.
- La prueba de que el total teórico no se filtra a la recepcionista pasa en la integración continua.
- Documentación actualizada.
