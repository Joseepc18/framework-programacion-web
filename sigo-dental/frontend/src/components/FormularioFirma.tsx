import { useState, type FormEvent } from 'react';
import { useAccion } from '../hooks/useAccion';
import { Aviso, Boton, Campo } from './ui';

interface Props {
  // Qué se firma ("esta nota", "la receta R-000001"): se muestra para que la firma sea consciente.
  descripcion: string;
  // Hace la firma con la contraseña; debe lanzar si falla (el error se muestra aquí).
  alFirmar: (password: string) => Promise<unknown>;
  alCancelar: () => void;
  textoBoton?: string;
}

// Reautenticación para firmar (firma digital): el profesional vuelve a escribir su contraseña.
const FormularioFirma = ({ descripcion, alFirmar, alCancelar, textoBoton = 'Firmar' }: Props) => {
  const accion = useAccion();
  const [password, setPassword] = useState('');

  const enviar = async (evento: FormEvent) => {
    evento.preventDefault();
    if (await accion.ejecutar(() => alFirmar(password))) setPassword('');
  };

  return (
    <form onSubmit={enviar} className="space-y-3 rounded-control border border-vidrio-borde p-4">
      <p className="text-etiqueta">Vas a firmar {descripcion}. La firma guarda tu identidad, la fecha y la huella del contenido, y lo bloquea.</p>
      <Campo id={`firma-${descripcion}`} etiqueta="Tu contraseña" type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} />
      {accion.error && <Aviso>{accion.error}</Aviso>}
      <div className="flex flex-wrap gap-3">
        <Boton type="submit" disabled={accion.ocupado || !password}>{accion.ocupado ? 'Firmando…' : textoBoton}</Boton>
        <Boton variante="secundario" onClick={alCancelar}>Cancelar</Boton>
      </div>
    </form>
  );
};

export default FormularioFirma;
