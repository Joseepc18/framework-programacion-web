import { nombreCompleto, type PacienteDetalle } from '../services/pacienteService';
import AlertasMedicas from './AlertasMedicas';
import { Chip } from './ui';

// Cabecera del expediente: identifica al paciente y mantiene visible la alerta médica (RF-02).
// Las pantallas clínicas de las siguientes fases la reutilizan.
const CabeceraPaciente = ({ paciente }: { paciente: PacienteDetalle }) => (
  <header className="vidrio space-y-4 rounded-tarjeta p-5 md:p-6">
    <div className="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h1 className="text-pagina font-semibold">{nombreCompleto(paciente)}</h1>
        <p className="mt-1 text-texto-suave">
          <span className="font-mono">{paciente.numero_historia}</span> · <span className="font-mono">{paciente.edad}</span> años ·{' '}
          <span className="font-mono">{paciente.documento}</span>
        </p>
      </div>
      <div className="flex flex-wrap gap-2">
        <Chip tono={paciente.activo ? 'acento' : 'neutro'}>{paciente.activo ? 'Activo' : 'Desactivado'}</Chip>
        <Chip>{paciente.convenio_nombre || 'Sin convenio'}</Chip>
      </div>
    </div>
    <AlertasMedicas alertas={paciente.alertas} />
  </header>
);

export default CabeceraPaciente;
