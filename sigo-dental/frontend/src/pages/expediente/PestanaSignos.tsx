import { useState, type FormEvent } from 'react';
import { Aviso, Boton, Campo, Tarjeta, Td, Tabla, Th } from '../../components/ui';
import { useAccion } from '../../hooks/useAccion';
import { useCarga } from '../../hooks/useCarga';
import { listarSignos, registrarSignos, type DatosSignos } from '../../services/expedienteService';
import { fechaHoraLocal } from '../../utils/fechas';

type Campos = Record<'sistolica' | 'diastolica' | 'frecuencia_cardiaca' | 'frecuencia_respiratoria' | 'temperatura' | 'saturacion' | 'peso_kg' | 'talla_cm' | 'observaciones', string>;

const VACIO: Campos = {
  sistolica: '', diastolica: '', frecuencia_cardiaca: '', frecuencia_respiratoria: '', temperatura: '',
  saturacion: '', peso_kg: '', talla_cm: '', observaciones: '',
};

// Convierte el texto del formulario: vacío = sin medición; coma decimal admitida.
const numero = (texto: string) => (texto.trim() === '' ? null : Number(texto.replace(',', '.')));

const mostrar = (valor: number | null, unidad = '') => (valor === null ? '—' : `${valor}${unidad}`);

// Signos vitales por visita (los registra el asistente o el odontólogo). Cada toma es un registro nuevo.
const PestanaSignos = ({ pacienteId }: { pacienteId: number }) => {
  const signos = useCarga(() => listarSignos(pacienteId), `signos-${pacienteId}`);
  const accion = useAccion();
  const [campos, setCampos] = useState<Campos>(VACIO);
  const poner = (campo: keyof Campos) => (e: { target: { value: string } }) => setCampos({ ...campos, [campo]: e.target.value });

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    const datos: DatosSignos = {
      sistolica: numero(campos.sistolica), diastolica: numero(campos.diastolica),
      frecuencia_cardiaca: numero(campos.frecuencia_cardiaca), frecuencia_respiratoria: numero(campos.frecuencia_respiratoria),
      temperatura: numero(campos.temperatura), saturacion: numero(campos.saturacion),
      peso_kg: numero(campos.peso_kg), talla_cm: numero(campos.talla_cm), observaciones: campos.observaciones,
    };
    if (await accion.ejecutar(() => registrarSignos(pacienteId, datos))) {
      setCampos(VACIO);
      signos.recargar();
    }
  };

  return (
    <div className="space-y-6">
      <Tarjeta>
        <form onSubmit={guardar} className="space-y-5">
          <h2 className="text-seccion font-semibold">Registrar signos vitales</h2>
          <div className="grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
            <Campo id="sv-sistolica" etiqueta="Presión sistólica (mmHg)" inputMode="numeric" value={campos.sistolica} onChange={poner('sistolica')} />
            <Campo id="sv-diastolica" etiqueta="Presión diastólica (mmHg)" inputMode="numeric" value={campos.diastolica} onChange={poner('diastolica')} />
            <Campo id="sv-fc" etiqueta="Frecuencia cardiaca (lpm)" inputMode="numeric" value={campos.frecuencia_cardiaca} onChange={poner('frecuencia_cardiaca')} />
            <Campo id="sv-fr" etiqueta="Frecuencia respiratoria (rpm)" inputMode="numeric" value={campos.frecuencia_respiratoria} onChange={poner('frecuencia_respiratoria')} />
            <Campo id="sv-temp" etiqueta="Temperatura (°C)" inputMode="decimal" value={campos.temperatura} onChange={poner('temperatura')} />
            <Campo id="sv-sat" etiqueta="Saturación de oxígeno (%)" inputMode="numeric" value={campos.saturacion} onChange={poner('saturacion')} />
            <Campo id="sv-peso" etiqueta="Peso (kg)" inputMode="decimal" value={campos.peso_kg} onChange={poner('peso_kg')} />
            <Campo id="sv-talla" etiqueta="Talla (cm)" inputMode="decimal" value={campos.talla_cm} onChange={poner('talla_cm')} />
          </div>
          <Campo id="sv-obs" etiqueta="Observaciones (opcional)" maxLength={300} value={campos.observaciones} onChange={poner('observaciones')} />
          {accion.error && <Aviso>{accion.error}</Aviso>}
          <Boton type="submit" disabled={accion.ocupado}>{accion.ocupado ? 'Guardando…' : 'Guardar toma'}</Boton>
        </form>
      </Tarjeta>

      {signos.error && <Aviso>{signos.error}</Aviso>}
      {signos.datos && signos.datos.length === 0 && <p className="text-texto-suave">Aún no hay signos vitales registrados.</p>}
      {signos.datos && signos.datos.length > 0 && (
        <Tabla etiqueta="Historial de signos vitales">
          <thead>
            <tr>
              <Th>Fecha</Th><Th>Presión</Th><Th>FC</Th><Th>FR</Th><Th>Temp.</Th><Th>SatO₂</Th><Th>Peso</Th><Th>Talla</Th><Th>Registró</Th>
            </tr>
          </thead>
          <tbody>
            {signos.datos.map((s) => (
              <tr key={s.id}>
                <Td className="font-mono text-etiqueta">{fechaHoraLocal(s.registrado_en)}</Td>
                <Td className="font-mono">{s.sistolica === null ? '—' : `${s.sistolica}/${s.diastolica}`}</Td>
                <Td className="font-mono">{mostrar(s.frecuencia_cardiaca)}</Td>
                <Td className="font-mono">{mostrar(s.frecuencia_respiratoria)}</Td>
                <Td className="font-mono">{mostrar(s.temperatura, ' °C')}</Td>
                <Td className="font-mono">{mostrar(s.saturacion, ' %')}</Td>
                <Td className="font-mono">{mostrar(s.peso_kg, ' kg')}</Td>
                <Td className="font-mono">{mostrar(s.talla_cm, ' cm')}</Td>
                <Td>{s.registrado_por}{s.observaciones && <span className="block text-etiqueta text-texto-suave">{s.observaciones}</span>}</Td>
              </tr>
            ))}
          </tbody>
        </Tabla>
      )}
    </div>
  );
};

export default PestanaSignos;
