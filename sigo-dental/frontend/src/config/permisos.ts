// Roles y permisos de la interfaz. Los códigos coinciden con roles.codigo del backend.
// Esto solo decide qué se muestra: la autoridad real es el backend (RequireRoles, responde 403).
import {
  Activity,
  Receipt,
  CalendarDays,
  ClipboardList,
  FileSearch,
  HandHelping,
  History,
  House,
  Library,
  Shield,
  Smile,
  Stethoscope,
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
  // Roles que pueden entrar; sin esta lista, cualquier cuenta con sesión.
  roles?: readonly Rol[];
}

// Cada rol tiene un área de prueba que consulta la ruta GET /api/prueba/<slug> del backend,
// la cual solo acepta a ese rol. Se reemplazan por los módulos reales cuando existan.
export interface AreaRol {
  rol: Rol;
  slug: string;
  label: string;
  icon: LucideIcon;
}

export const AREAS: readonly AreaRol[] = [
  { rol: 'admin_ti', slug: 'admin-ti', label: 'Área de administración TI', icon: Shield },
  { rol: 'recepcionista', slug: 'recepcion', label: 'Área de recepción', icon: ClipboardList },
  { rol: 'odontologo', slug: 'odontologo', label: 'Área del odontólogo', icon: Stethoscope },
  { rol: 'asistente', slug: 'asistente', label: 'Área del asistente', icon: HandHelping },
  { rol: 'auditor', slug: 'auditor', label: 'Área de auditoría', icon: FileSearch },
];

export const rutaDeArea = (area: AreaRol) => `/area/${area.slug}`;

export const MENU: readonly ElementoMenu[] = [
  { to: '/', label: 'Inicio', icon: House },
  { to: '/portal', label: 'Mi portal', icon: Smile, roles: ['paciente'] },
  { to: '/estado', label: 'Estado del sistema', icon: Activity, roles: ['admin_ti'] },
  { to: '/catalogos', label: 'Catálogos', icon: Library, roles: ['admin_ti'] },
  { to: '/pacientes', label: 'Pacientes', icon: Users, roles: ['recepcionista', 'odontologo', 'asistente'] },
  { to: '/auditoria', label: 'Auditoría', icon: History, roles: ['admin_ti', 'auditor'] },
  { to: '/cobros', label: 'Cobros', icon: Receipt, roles: ['recepcionista', 'auditor'] },
  { to: '/agenda', label: 'Agenda', icon: CalendarDays, roles: ['recepcionista', 'odontologo', 'asistente'] },
  ...AREAS.map((area) => ({ to: rutaDeArea(area), label: area.label, icon: area.icon, roles: [area.rol] })),
];

export const puedeVer = (elemento: Pick<ElementoMenu, 'roles'>, rol: Rol) =>
  !elemento.roles || elemento.roles.includes(rol);

export const menuPara = (rol: Rol) => MENU.filter((elemento) => puedeVer(elemento, rol));
