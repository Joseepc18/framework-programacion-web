// Estados de la cita y sus transiciones. Reflejan las reglas del backend (models/cita.go), que es la autoridad:
// aquí solo se decide qué botones mostrar.
import type { Rol } from './permisos';
import type { EstadoCita } from '../services/citaService';

export const ETIQUETA_ESTADO: Record<EstadoCita, string> = {
  programado: 'Programado',
  confirmado: 'Confirmado',
  en_sala: 'En sala de espera',
  en_atencion: 'En atención',
  atendido: 'Atendido',
  cancelado: 'Cancelado',
  no_asistio: 'No asistió',
};

// Clases completas (no armadas por partes) para que Tailwind las detecte; los colores son tokens de index.css.
export const ESTILO_ESTADO: Record<EstadoCita, string> = {
  programado: 'border-cita-programado-borde bg-cita-programado-fondo text-cita-programado-texto',
  confirmado: 'border-cita-confirmado-borde bg-cita-confirmado-fondo text-cita-confirmado-texto',
  en_sala: 'border-cita-en-sala-borde bg-cita-en-sala-fondo text-cita-en-sala-texto',
  en_atencion: 'border-cita-en-atencion-borde bg-cita-en-atencion-fondo text-cita-en-atencion-texto',
  atendido: 'border-cita-atendido-borde bg-cita-atendido-fondo text-cita-atendido-texto',
  cancelado: 'border-dashed border-cita-cancelado-borde bg-cita-cancelado-fondo text-cita-cancelado-texto line-through',
  no_asistio: 'border-cita-no-asistio-borde bg-cita-no-asistio-fondo text-cita-no-asistio-texto',
};

const TRANSICIONES: Record<EstadoCita, readonly EstadoCita[]> = {
  programado: ['confirmado', 'en_sala', 'cancelado', 'no_asistio'],
  confirmado: ['en_sala', 'cancelado', 'no_asistio'],
  en_sala: ['en_atencion', 'cancelado'],
  en_atencion: ['atendido'],
  atendido: [],
  cancelado: [],
  no_asistio: [],
};

export const ETIQUETA_ACCION: Record<EstadoCita, string> = {
  programado: 'Volver a programado',
  confirmado: 'Confirmar',
  en_sala: 'Llegó a la sala de espera',
  en_atencion: 'Iniciar atención',
  atendido: 'Marcar como atendido',
  cancelado: 'Cancelar cita',
  no_asistio: 'No asistió',
};

// Estados a los que el rol puede pasar la cita: la recepción, a cualquiera permitido; el odontólogo y el asistente,
// solo a En atención y Atendido; el resto, a ninguno.
export function estadosSiguientes(estado: EstadoCita, rol: Rol): readonly EstadoCita[] {
  const posibles = TRANSICIONES[estado];
  if (rol === 'recepcionista') return posibles;
  if (rol === 'odontologo' || rol === 'asistente') return posibles.filter((e) => e === 'en_atencion' || e === 'atendido');
  return [];
}

// Una cita se puede reprogramar mientras no haya empezado.
export const sePuedeReprogramar = (estado: EstadoCita) => estado === 'programado' || estado === 'confirmado';

// Las citas canceladas o con inasistencia no ocupan el sillón.
export const liberaSillon = (estado: EstadoCita) => estado === 'cancelado' || estado === 'no_asistio';
