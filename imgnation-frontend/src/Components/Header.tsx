import React from 'react';
import { User, Search } from 'lucide-react';

interface HeaderProps {
  searchValue: string;
  onSearchChange: (value: string) => void;
  onSearch: () => void;
  onAccountClick: () => void;
  searchPlaceholder?: string;
  className?: string;
}

interface AccountButtonProps {
  onClick: () => void;
  className?: string;
}

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
  placeholder?: string;
  initValue?: string;
  className?: string;
}

export const SearchBar: React.FC<SearchBarProps> = ({
  value,
  onChange,
  onSubmit,
  placeholder = "Search...",
  className = ""
}) => {
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      onSubmit();
    }
  };

  return (
    <div className={`flex-1 relative ${className}`}>
      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
      <input
        type="text"
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={handleKeyPress}
        className="w-full bg-gray-200 border rounded-full py-3 pl-12 pr-4 text-gray-700 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
      />
    </div>
  );
};


export const AccountButton: React.FC<AccountButtonProps> = ({
  onClick,
  className = ""
}) => {
  const baseClasses = "flex items-center justify-center transition-transform hover:scale-105";
  const variantClasses = "w-12 h-12 bg-gradient-to-r from-purple-500 to-pink-500 text-white rounded-full"

  return (
    <button
      type="button"
      onClick={onClick}
      className={`${baseClasses} ${variantClasses} ${className}`}
    >
      <User className="w-6 h-6 text-white" />
    </button>
  );
};


export const Header: React.FC<HeaderProps> = ({
  searchValue,
  onSearchChange,
  onSearch,
  onAccountClick,
  searchPlaceholder = "Search tags...",
  className = ""
}) => {
  return (
    <header className={`sticky top-0 z-50 bg-gray-100 backdrop-blur-lg border-b shadow-lg ${className}`}>
      <div className="max-w-6xl mx-auto px-4 py-4">
        <div className="flex items-center gap-4">
          <SearchBar
            value={searchValue}
            onChange={onSearchChange}
            onSubmit={onSearch}
            placeholder={searchPlaceholder}
          />

          <AccountButton
            onClick={onAccountClick}
          />
        </div>
      </div>
    </header>
  );
};