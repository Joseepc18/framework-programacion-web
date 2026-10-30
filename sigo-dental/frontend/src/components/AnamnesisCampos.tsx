import { Casilla, CampoTexto } from './ui';
import type { Anamnesis } from '../services/pacienteService';
import type { ItemCatalogo } from '../services/catalogoService';

interface AnamnesisCamposProps {
  valor: Anamnesis;
  onChange: (valor: Anamnesis) => void;
  condiciones: ItemCatalogo[];
  prefijo: string;
}

// Los cuatro apartados son obligatorios; si no hay nada que declarar se escribe "Ninguna".
const AnamnesisCampos = ({ valor, onChange, condiciones, prefijo }: AnamnesisCamposProps) => {
  const cambiar = (campo: keyof Omit<Anamnesis, 'condiciones_criticas'>, texto: string) => onChange({ ...valor, [campo]: texto });

  const alternar = (id: number, marcada: boolean) =>
    onChange({
      ...valor,
      condiciones_criticas: marcada ? [...valor.condiciones_criticas, id] : valor.condiciones_criticas.filter((c) => c !== id),
    });

  return (
    <div className="space-y-5">
      <p className="text-etiqueta text-texto-suave">Todos los apartados son obligatorios. Si no hay nada que declarar, escribe "Ninguna".</p>
      <div className="grid gap-5 md:grid-cols-2">
        <CampoTexto id={`${prefijo}-sistemicas`} etiqueta="Enfermedades sistémicas" required maxLength={2000} value={valor.enfermedades_sistemicas} onChange={(e) => cambiar('enfermedades_sistemicas', e.target.value)} />
        <CampoTexto id={`${prefijo}-alergias`} etiqueta="Alergias a fármacos" required maxLength={2000} value={valor.alergias_farmacos} onChange={(e) => cambiar('alergias_farmacos', e.target.value)} />
        <CampoTexto id={`${prefijo}-quirurgicos`} etiqueta="Antecedentes quirúrgicos" required maxLength={2000} value={valor.antecedentes_quirurgicos} onChange={(e) => cambiar('antecedentes_quirurgicos', e.target.value)} />
        <CampoTexto id={`${prefijo}-coagulopatias`} etiqueta="Coagulopatías" required maxLength={2000} value={valor.coagulopatias} onChange={(e) => cambiar('coagulopatias', e.target.value)} />
      </div>
      <fieldset>
        <legend className="text-etiqueta font-medium text-texto-suave">Condiciones críticas (activan la alerta médica en el expediente)</legend>
        <div className="mt-2 grid gap-x-6 sm:grid-cols-2 xl:grid-cols-3">
          {condiciones.map((c) => (
            <Casilla key={c.id} id={`${prefijo}-cc-${c.id}`} etiqueta={c.nombre} checked={valor.condiciones_criticas.includes(c.id)} onChange={(e) => alternar(c.id, e.target.checked)} />
          ))}
        </div>
      </fieldset>
    </div>
  );
};

export default AnamnesisCampos;
