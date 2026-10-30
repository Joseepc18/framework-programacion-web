import { useState, type FormEvent } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import AnamnesisCampos from '../components/AnamnesisCampos';
import { Aviso, Boton, BotonEnlace, Campo, Casilla, Selector, Tarjeta, Titulo } from '../components/ui';
import { useAccion } from '../hooks/useAccion';
import { useCarga } from '../hooks/useCarga';
import { listarSimples, type ItemCatalogo } from '../services/catalogoService';
import { actualizarPaciente, ANAMNESIS_VACIA, crearPaciente, obtenerPaciente, type Anamnesis, type DatosPaciente } from '../services/pacienteService';
import { edadDe, hoyISO } from '../utils/fechas';

const VACIO: DatosPaciente = {
  documento: '', nombres: '', apellidos: '', fecha_nacimiento: '', sexo: 'femenino', telefono: '', email: '', direccion: '',
  emergencia_nombre: '', emergencia_telefono: '', emergencia_parentesco: '',
  tutor_nombre: '', tutor_documento: '', tutor_telefono: '',
  convenio_id: null, canal_preferido: 'whatsapp', activo: true,
};

interface FormularioProps {
  inicial: DatosPaciente;
  id?: number;
  convenios: ItemCatalogo[];
  condiciones: ItemCatalogo[];
}

const Formulario = ({ inicial, id, convenios, condiciones }: FormularioProps) => {
  const navegar = useNavigate();
  const accion = useAccion();
  const [datos, setDatos] = useState(inicial);
  const [anamnesis, setAnamnesis] = useState<Anamnesis>(ANAMNESIS_VACIA);

  const cambiar = <K extends keyof DatosPaciente>(campo: K, valor: DatosPaciente[K]) => setDatos((d) => ({ ...d, [campo]: valor }));
  const texto = (campo: keyof DatosPaciente) => ({ value: String(datos[campo] ?? ''), onChange: (e: { target: { value: string } }) => cambiar(campo, e.target.value as never) });

  const edad = edadDe(datos.fecha_nacimiento);
  const esMenor = edad !== null && edad < 18;

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    let destino = '';
    const bien = await accion.ejecutar(async () => {
      const paciente = id ? await actualizarPaciente(id, datos) : await crearPaciente(datos, anamnesis);
      destino = `/pacientes/${paciente.id}`;
    });
    if (bien) navegar(destino);
  };

  return (
    <form onSubmit={guardar} className="space-y-6">
      <Tarjeta>
        <h2 className="text-seccion font-semibold">Datos personales</h2>
        <div className="mt-5 grid gap-5 md:grid-cols-2 xl:grid-cols-3">
          <Campo id="documento" etiqueta="Documento de identidad" required maxLength={30} {...texto('documento')} />
          <Campo id="nombres" etiqueta="Nombres" required maxLength={100} {...texto('nombres')} />
          <Campo id="apellidos" etiqueta="Apellidos" required maxLength={100} {...texto('apellidos')} />
          <Campo id="nacimiento" etiqueta="Fecha de nacimiento" type="date" required min="1900-01-01" max={hoyISO()} {...texto('fecha_nacimiento')} ayuda={edad !== null ? `${edad} años` : undefined} />
          <Selector id="sexo" etiqueta="Sexo" required value={datos.sexo} onChange={(e) => cambiar('sexo', e.target.value as DatosPaciente['sexo'])}>
            <option value="femenino">Femenino</option>
            <option value="masculino">Masculino</option>
            <option value="otro">Otro</option>
          </Selector>
          <Campo id="telefono" etiqueta="Teléfono" type="tel" required maxLength={30} {...texto('telefono')} />
          <Campo id="email" etiqueta="Correo electrónico (opcional)" type="email" maxLength={150} {...texto('email')} />
          <Campo id="direccion" etiqueta="Dirección (opcional)" maxLength={200} {...texto('direccion')} />
          <Selector id="canal" etiqueta="Canal preferido para avisos" value={datos.canal_preferido} onChange={(e) => cambiar('canal_preferido', e.target.value as DatosPaciente['canal_preferido'])}>
            <option value="whatsapp">WhatsApp</option>
            <option value="sms">SMS</option>
            <option value="email">Correo electrónico</option>
          </Selector>
          <Selector id="convenio" etiqueta="Seguro dental o convenio" value={datos.convenio_id ?? ''} onChange={(e) => cambiar('convenio_id', e.target.value ? Number(e.target.value) : null)}>
            <option value="">Sin convenio</option>
            {convenios.map((c) => <option key={c.id} value={c.id}>{c.nombre}</option>)}
          </Selector>
        </div>
      </Tarjeta>

      <Tarjeta>
        <h2 className="text-seccion font-semibold">Contacto de emergencia</h2>
        <div className="mt-5 grid gap-5 md:grid-cols-3">
          <Campo id="emergencia-nombre" etiqueta="Nombre" required maxLength={150} {...texto('emergencia_nombre')} />
          <Campo id="emergencia-telefono" etiqueta="Teléfono" type="tel" required maxLength={30} {...texto('emergencia_telefono')} />
          <Campo id="emergencia-parentesco" etiqueta="Parentesco (opcional)" maxLength={50} {...texto('emergencia_parentesco')} />
        </div>
      </Tarjeta>

      <Tarjeta>
        <h2 className="text-seccion font-semibold">Tutor legal</h2>
        <p className="mt-1 text-etiqueta text-texto-suave">{esMenor ? 'Obligatorio: el paciente es menor de 18 años.' : 'Solo es obligatorio para menores de 18 años.'}</p>
        <div className="mt-5 grid gap-5 md:grid-cols-3">
          <Campo id="tutor-nombre" etiqueta="Nombre" required={esMenor} maxLength={150} {...texto('tutor_nombre')} />
          <Campo id="tutor-documento" etiqueta="Documento" required={esMenor} maxLength={30} {...texto('tutor_documento')} />
          <Campo id="tutor-telefono" etiqueta="Teléfono" type="tel" required={esMenor} maxLength={30} {...texto('tutor_telefono')} />
        </div>
      </Tarjeta>

      {id ? (
        <Tarjeta><Casilla id="activo" etiqueta="Paciente activo" checked={datos.activo} onChange={(e) => cambiar('activo', e.target.checked)} /></Tarjeta>
      ) : (
        <Tarjeta>
          <h2 className="text-seccion font-semibold">Anamnesis</h2>
          <div className="mt-5"><AnamnesisCampos prefijo="alta" valor={anamnesis} onChange={setAnamnesis} condiciones={condiciones} /></div>
        </Tarjeta>
      )}

      {accion.error && <Aviso>{accion.error}</Aviso>}
      <div className="flex flex-wrap gap-3">
        <Boton type="submit" disabled={accion.ocupado}>{accion.ocupado ? 'Guardando…' : id ? 'Guardar cambios' : 'Registrar paciente'}</Boton>
        <BotonEnlace variante="secundario" to={id ? `/pacientes/${id}` : '/pacientes'}>Cancelar</BotonEnlace>
      </div>
    </form>
  );
};

// Alta de un paciente (con su anamnesis) o edición de su ficha; el formulario se monta cuando ya hay datos.
const PacienteFormPage = () => {
  const { id: idRuta } = useParams();
  const id = idRuta ? Number(idRuta) : undefined;
  const paciente = useCarga(() => (id ? obtenerPaciente(id) : Promise.resolve(null)), `paciente-${id ?? 'nuevo'}`);
  const convenios = useCarga(() => listarSimples('convenios'), 'convenios-activos');
  const condiciones = useCarga(() => listarSimples('condiciones-criticas'), 'condiciones-activas');

  const error = paciente.error || convenios.error || condiciones.error;
  const listo = !paciente.cargando && convenios.datos && condiciones.datos && (!id || paciente.datos);

  return (
    <section aria-labelledby="titulo-form-paciente" className="space-y-6">
      <Titulo id="titulo-form-paciente">{id ? 'Editar paciente' : 'Nuevo paciente'}</Titulo>
      {error && <Aviso>{error}</Aviso>}
      {!listo && !error && <p>Cargando…</p>}
      {listo && (
        <Formulario
          inicial={paciente.datos ?? VACIO}
          id={id}
          convenios={convenios.datos ?? []}
          condiciones={condiciones.datos ?? []}
        />
      )}
    </section>
  );
};

export default PacienteFormPage;
