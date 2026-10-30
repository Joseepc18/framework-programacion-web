// Roles y permisos de la interfaz. Los códigos coinciden con roles.codigo del backend.
// Esto solo decide qué se muestra: la autoridad real es el backend (RequireRoles, responde 403).
import {
  Activity,
  Receipt,
  CalendarDays,
  History,
  House,
  Library,
  Smile,
  Users,
  type LucideIcon,
} from 'lucide-react';

// 'paciente' es la cuenta del portal del paciente (F2-8); los otros cinco son los roles del personal.
export type Rol = 'admin_ti' | 'recepcionista' | 'odontologo' | 'asistente' | 'auditor' | 'paciente';

export const ETIQUETA_ROL: Record<Rol, string> = {
  admin_ti: 'Administrador TI',
  recepcionista: 'Recepcionista',
  odontologo: 'Odontólogo',
  asistente: 'Asistente dental',
  auditor: 'Auditor',
  paciente: 'Paciente',
};

export const esRol = (valor: unknown): valor is Rol =>
  typeof valor === 'string' && Object.hasOwn(ETIQUETA_ROL, valor);

export interface ElementoMenu {
  to: string;
  label: string;
  icon: LucideIcon;
  // Frase corta que usa la pantalla de Inicio como acceso directo.
  descripcion?: string;
  // Roles que pueden entrar; sin esta lista, cualquier cuenta con sesión.
  roles?: readonly Rol[];
}

export const MENU: readonly ElementoMenu[] = [
  { to: '/', label: 'Inicio', icon: House },
  { to: '/portal', label: 'Mi portal', icon: Smile, roles: ['paciente'], descripcion: 'Tus citas, recetas y consentimientos.' },
  { to: '/estado', label: 'Estado del sistema', icon: Activity, roles: ['admin_ti'], descripcion: 'Comprueba que la base de datos y la caché respondan.' },
  { to: '/catalogos', label: 'Catálogos', icon: Library, roles: ['admin_ti'], descripcion: 'Sillones, profesionales, procedimientos y parámetros de la clínica.' },
  { to: '/pacientes', label: 'Pacientes', icon: Users, roles: ['recepcionista', 'odontologo', 'asistente'], descripcion: 'Fichas, anamnesis y expediente clínico.' },
  { to: '/agenda', label: 'Agenda', icon: CalendarDays, roles: ['recepcionista', 'odontologo', 'asistente'], descripcion: 'Citas del día y de la semana por sillón.' },
  { to: '/cobros', label: 'Cobros', icon: Receipt, roles: ['recepcionista', 'auditor'], descripcion: 'Comprobantes, pagos y anulaciones.' },
  { to: '/auditoria', label: 'Auditoría', icon: History, roles: ['admin_ti', 'auditor'], descripcion: 'Quién consultó o cambió cada registro y cuándo.' },
];

export const puedeVer = (elemento: Pick<ElementoMenu, 'roles'>, rol: Rol) =>
  !elemento.roles || elemento.roles.includes(rol);

export const menuPara = (rol: Rol) => MENU.filter((elemento) => puedeVer(elemento, rol));
