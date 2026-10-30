import { useState, type FormEvent } from 'react';
import { Aviso, Boton, Campo, Chip, Selector, Tarjeta, Td, Tabla, Th } from '../../components/ui';
import { useAccion } from '../../hooks/useAccion';
import { mensajeDeApi, useCarga } from '../../hooks/useCarga';
import {
  descargarArchivo, ETIQUETA_CATEGORIA, listarArchivos, subirArchivo, type Archivo,
} from '../../services/expedienteService';
import { abrirBlob, tamanoLegible } from '../../utils/archivos';
import { fechaHoraLocal } from '../../utils/fechas';

const CATEGORIAS_SUBIBLES = ['radiografia', 'fotografia', 'documento'] as const;

// Archivos del paciente (radiografías, fotos y PDF). Se descargan con la sesión; no hay enlaces públicos.
const PestanaArchivos = ({ pacienteId }: { pacienteId: number }) => {
  const archivos = useCarga(() => listarArchivos(pacienteId), `archivos-${pacienteId}`);
  const accion = useAccion();
  const [archivo, setArchivo] = useState<File | null>(null);
  // Cambiar la clave vacía el campo de archivo tras subir.
  const [claveEntrada, setClaveEntrada] = useState(0);
  const [categoria, setCategoria] = useState<(typeof CATEGORIAS_SUBIBLES)[number]>('radiografia');
  const [descripcion, setDescripcion] = useState('');
  const [errorDescarga, setErrorDescarga] = useState('');

  const subir = async (evento: FormEvent) => {
    evento.preventDefault();
    if (!archivo) return;
    if (await accion.ejecutar(() => subirArchivo(pacienteId, categoria, descripcion, archivo))) {
      setDescripcion('');
      setArchivo(null);
      setClaveEntrada((n) => n + 1);
      archivos.recargar();
    }
  };

  const abrir = async (archivo: Archivo) => {
    setErrorDescarga('');
    try {
      abrirBlob(await descargarArchivo(archivo.id), archivo.nombre_original, !archivo.tipo_mime.startsWith('image/') && archivo.tipo_mime !== 'application/pdf');
    } catch (fallo) {
      setErrorDescarga(mensajeDeApi(fallo));
    }
  };

  return (
    <div className="space-y-6">
      <Tarjeta>
        <form onSubmit={subir} className="space-y-5">
          <h2 className="text-seccion font-semibold">Subir archivo</h2>
          <div className="grid gap-5 md:grid-cols-3">
            <Selector id="arch-categoria" etiqueta="Categoría" value={categoria} onChange={(e) => setCategoria(e.target.value as typeof categoria)}>
              {CATEGORIAS_SUBIBLES.map((c) => <option key={c} value={c}>{ETIQUETA_CATEGORIA[c]}</option>)}
            </Selector>
            <Campo id="arch-descripcion" etiqueta="Descripción (opcional)" maxLength={200} value={descripcion} onChange={(e) => setDescripcion(e.target.value)} />
            <Campo key={claveEntrada} id="arch-archivo" etiqueta="Archivo" type="file" required onChange={(e) => setArchivo(e.target.files?.[0] ?? null)} accept="image/jpeg,image/png,image/webp,application/pdf"
              ayuda="JPG, PNG, WebP o PDF" />
          </div>
          {accion.error && <Aviso>{accion.error}</Aviso>}
          <Boton type="submit" disabled={accion.ocupado}>{accion.ocupado ? 'Subiendo…' : 'Subir'}</Boton>
        </form>
      </Tarjeta>

      {errorDescarga && <Aviso>{errorDescarga}</Aviso>}
      {archivos.error && <Aviso>{archivos.error}</Aviso>}
      {archivos.datos && archivos.datos.length === 0 && <p className="text-texto-suave">El paciente aún no tiene archivos.</p>}
      {archivos.datos && archivos.datos.length > 0 && (
        <Tabla etiqueta="Archivos del paciente">
          <thead>
            <tr><Th>Archivo</Th><Th>Categoría</Th><Th>Tamaño</Th><Th>Subido</Th><Th><span className="sr-only">Acciones</span></Th></tr>
          </thead>
          <tbody>
            {archivos.datos.map((a) => (
              <tr key={a.id}>
                <Td>
                  {a.nombre_original}
                  {a.descripcion && <span className="block text-etiqueta text-texto-suave">{a.descripcion}</span>}
                </Td>
                <Td><Chip>{ETIQUETA_CATEGORIA[a.categoria]}</Chip></Td>
                <Td className="font-mono text-etiqueta">{tamanoLegible(a.tamano_bytes)}</Td>
                <Td className="text-etiqueta">{fechaHoraLocal(a.subido_en)}<span className="block text-texto-suave">{a.subido_por}</span></Td>
                <Td><Boton variante="secundario" onClick={() => abrir(a)}>Abrir</Boton></Td>
              </tr>
            ))}
          </tbody>
        </Tabla>
      )}
    </div>
  );
};

export default PestanaArchivos;
