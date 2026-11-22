import Search from "@react-spectrum/s2/icons/Search";
import { Input } from "@cueia/ui/input";

export function SearchBox({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <div className="relative [&>svg]:absolute [&>svg]:z-1 [&>svg]:left-2.75 [&>svg]:top-2.5 [&>svg]:w-4 [&>svg]:h-4 [&>svg]:pointer-events-none [&>svg]:text-muted [&_input]:ps-6.75">
      <Search aria-hidden="true" />
      <Input
        type="search"
        aria-label={label}
        placeholder="Search"
        value={value}
        onChange={onChange}
      />
    </div>
  );
}
