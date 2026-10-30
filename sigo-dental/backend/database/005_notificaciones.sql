-- Migración 005: bitácora de notificaciones de citas (Fase 1, F1-4). Requiere la migración 004.

-- El paciente puede confirmar su cita desde un enlace, sin cuenta: ese cambio de estado no tiene usuario.
ALTER TABLE cita_historial ALTER COLUMN usuario_id DROP NOT NULL;

CREATE TABLE IF NOT EXISTS notificaciones (
    id              INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    cita_id         INTEGER      NOT NULL REFERENCES citas (id),
    tipo            VARCHAR(20)  NOT NULL CHECK (tipo IN ('confirmacion', 'recordatorio_48h', 'recordatorio_24h')),
    canal           VARCHAR(20)  NOT NULL CHECK (canal IN ('whatsapp', 'sms', 'email')),
    destino         VARCHAR(150) NOT NULL DEFAULT '',
    mensaje         TEXT         NOT NULL,
    estado          VARCHAR(20)  NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'enviada', 'fallida')),
    detalle         VARCHAR(300) NOT NULL DEFAULT '',
    -- Una notificación deja de ser vigente cuando la cita se reprograma: queda en la bitácora y se generan otras.
    vigente         BOOLEAN      NOT NULL DEFAULT TRUE,
    creado_en       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    enviado_en      TIMESTAMPTZ
);

-- Clave única por cita y tipo: reservar la fila antes de enviar garantiza un solo mensaje por tipo,
-- aunque el planificador se ejecute varias veces o haya varias instancias de la API.
CREATE UNIQUE INDEX IF NOT EXISTS uq_notificaciones_cita_tipo ON notificaciones (cita_id, tipo) WHERE vigente;
CREATE INDEX IF NOT EXISTS idx_notificaciones_cita ON notificaciones (cita_id, id);
