import { Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { useState } from "react";

interface MatchInputProps {
  label: string;
  items: string[];
  placeholder?: string;
  onChange: (items: string[]) => void;
}

export function MatchInput({
  label,
  items,
  placeholder,
  onChange,
}: MatchInputProps) {
  const [value, setValue] = useState("");

  const handleAdd = () => {
    if (value.trim()) {
      onChange([...items, value.trim()]);
      setValue("");
    }
  };

  const handleRemove = (item: string) => {
    onChange(items.filter((i) => i !== item));
  };

  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      <div className="flex items-center gap-2">
        <Input
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && value.trim()) {
              e.preventDefault();
              handleAdd();
            }
          }}
          placeholder={placeholder}
          className="rounded-xl"
        />
        <Button
          size="icon"
          variant="outline"
          onClick={handleAdd}
          className="rounded-xl"
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      <div className="flex flex-wrap gap-2 mt-2">
        {items.map((item, index) => (
          <Badge
            key={index}
            variant="secondary"
            className="flex items-center gap-1 rounded-full bg-secondary/50 px-3"
          >
            {item}
            <Button
              size="icon"
              variant="ghost"
              className="h-4 w-4 p-0 hover:bg-transparent"
              onClick={() => handleRemove(item)}
            >
              <X className="h-3 w-3" />
            </Button>
          </Badge>
        ))}
      </div>
    </div>
  );
}
