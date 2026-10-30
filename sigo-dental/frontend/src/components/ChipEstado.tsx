import { ESTILO_ESTADO, ETIQUETA_ESTADO } from '../config/citas';
import type { EstadoCita } from '../services/citaService';

// El estado siempre se escribe con su nombre; el color y el tachado solo lo refuerzan.
const ChipEstado = ({ estado }: { estado: EstadoCita }) => (
  <span className={`inline-block whitespace-nowrap rounded-full border px-3 py-1 text-etiqueta font-medium ${ESTILO_ESTADO[estado]}`}>
    {ETIQUETA_ESTADO[estado]}
  </span>
);

export default ChipEstado;
