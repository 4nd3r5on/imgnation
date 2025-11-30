import React, { useState, useRef, DragEvent, ChangeEvent, KeyboardEvent } from 'react';
import {
  Upload,
  X,
  File,
  Image,
  Video,
  FileText,
  Tag,
  Plus,
  Globe,
  Lock,
  Users,
  Mail,
  LucideProps
} from 'lucide-react';
import { useNavigate } from 'react-router';
import {useAuth} from "../../Providers/Api.tsx";

type UploadVisibilityT = 'anyone' | 'authorized' | 'private'

// Types
interface FileMetadata {
  tags: string[];
  visibility: UploadVisibilityT;
  emails: string[];
  description: string;
}

interface FileWithMetadata {
  file: File;
  metadata: FileMetadata;
}

interface PrivacyPolicy {
  value: UploadVisibilityT
  icon: React.ForwardRefExoticComponent<Omit<LucideProps, "ref"> & React.RefAttributes<SVGSVGElement>>
  label: string
}

const PrivacyPolicies: PrivacyPolicy[] = [
  { value: 'anyone', icon: Globe, label: 'Public' },
  { value: 'authorized', icon: Users, label: 'Authorized' },
  { value: 'private', icon: Lock, label: 'Private' }
]

interface FilePreviewProps {
  file: File;
  metadata: FileMetadata;
  onRemove: (file: File) => void;
  onUpdateMetadata: (file: File, metadata: FileMetadata) => void;
}

// File type detection utilities
const getFileIcon = (mimeType: string) => {
  if (mimeType.startsWith('image/')) return Image;
  if (mimeType.startsWith('video/')) return Video;
  if (mimeType.includes('text') || mimeType.includes('document')) return FileText;
  return File;
};

const isPreviewable = (mimeType: string): boolean => {
  return mimeType.startsWith('image/') || mimeType.startsWith('video/');
};

const formatFileSize = (bytes: number): string => {
  return (bytes / (1024 * 1024)).toFixed(2);
};

// Individual file preview component
const FilePreview: React.FC<FilePreviewProps> = ({ file, metadata, onRemove, onUpdateMetadata }) => {
  const [tagInput, setTagInput] = useState<string>('');
  const [emailInput, setEmailInput] = useState<string>('');
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);

  React.useEffect(() => {
    if (isPreviewable(file.type)) {
      const url = URL.createObjectURL(file);
      setPreviewUrl(url);
      return () => URL.revokeObjectURL(url);
    }
  }, [file]);

  const addTag = (): void => {
    const inputTags = tagInput.split(' ')
    const addTags: string[] = []
    for (const tag of inputTags) {
      if (tag.length < 1) continue
      if (metadata.tags.includes(tag)) continue
      addTags.push(tag)
    }
    const newTags = [...metadata.tags, ...addTags];
    setTagInput('');
    onUpdateMetadata(file, { ...metadata, tags: newTags });
  };

  const removeTag = (tagToRemove: string): void => {
    const newTags = metadata.tags.filter((tag: string) => tag !== tagToRemove);
    onUpdateMetadata(file, { ...metadata, tags: newTags });
  };

  const addEmail = (): void => {
    const inputEmails = emailInput.split(' ')
    const addEmails: string[] = []
    for (const email of inputEmails) {
      if (!email.includes('@') || email.length < 1) continue
      if (metadata.emails.includes(email)) continue
      addEmails.push(email)
    }
    const newEmails = [...metadata.emails, ...addEmails];
    setEmailInput('');
    onUpdateMetadata(file, { ...metadata, emails: newEmails });
  };

  const removeEmail = (emailToRemove: string): void => {
    const newEmails = metadata.emails.filter((email: string) => email !== emailToRemove);
    onUpdateMetadata(file, { ...metadata, emails: newEmails });
  };

  const handleVisibilityChange = (newVisibility: UploadVisibilityT): void => {
    onUpdateMetadata(file, { ...metadata, visibility: newVisibility });
  };

  const handleDescriptionChange = (description: string): void => {
    onUpdateMetadata(file, { ...metadata, description });
  };

  const FileIcon = getFileIcon(file.type);

  return (
    <div className="border border-gray-200 rounded-lg p-4 bg-white">
      <div className="flex gap-4">
        {/* Preview/Icon */}
        <div className="w-16 h-16 flex-shrink-0 bg-gray-100 rounded-lg flex items-center justify-center overflow-hidden">
          {previewUrl ? (
            file.type.startsWith('image/') ? (
              <img src={previewUrl} alt={file.name} className="w-full h-full object-cover" />
            ) : (
              <video src={previewUrl} className="w-full h-full object-cover" muted />
            )
          ) : (
            <FileIcon className="w-8 h-8 text-gray-400" />
          )}
        </div>

        {/* File Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between">
            <div className="flex-1 min-w-0">
              <h4 className="font-medium text-gray-900 truncate">{file.name}</h4>
              <p className="text-sm text-gray-500">{formatFileSize(file.size)} MB</p>
              <p className="text-xs text-gray-400">{file.type}</p>
            </div>
            <button
              type="button"
              onClick={() => onRemove(file)}
              className="ml-2 p-1 text-gray-400 hover:text-red-500 transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Description Section */}
      <div className="mt-4 space-y-3">
        <div className="flex items-center gap-2">
          <FileText className="w-4 h-4 text-gray-500" />
          <span className="text-sm font-medium text-gray-700">Description</span>
        </div>
        <textarea
          value={metadata.description}
          onChange={(e: ChangeEvent<HTMLTextAreaElement>) => handleDescriptionChange(e.target.value)}
          placeholder="Add a description for this file..."
          className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
          rows={2}
        />

        {/* Tags Section */}
        <div className="flex items-center gap-2">
          <Tag className="w-4 h-4 text-gray-500" />
          <span className="text-sm font-medium text-gray-700">Tags</span>
        </div>

        {/* Tags Input */}
        <div className="flex gap-2">
          <input
            type="text"
            value={tagInput}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setTagInput(e.target.value)}
            onKeyPress={(e: KeyboardEvent<HTMLInputElement>) => e.key === 'Enter' && addTag()}
            placeholder="Add tags (space-separated)"
            className="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            onClick={addTag}
            className="px-3 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors"
          >
            <Plus className="w-4 h-4" />
          </button>
        </div>

        {/* Tags List */}
        {metadata.tags.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {metadata.tags.map((tag, index) => (
              <span
                key={index}
                onClick={() => removeTag(tag)}
                className="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full cursor-pointer hover:bg-blue-200 transition-colors max-w-32 truncate"
              >
                <span className="truncate">{tag}</span>
                <X className="w-3 h-3 ml-1 flex-shrink-0" />
              </span>
            ))}
          </div>
        )}

        {/* Visibility Toggle */}
        <div className="space-y-2">
          <span className="text-sm font-medium text-gray-700">Visibility</span>
          <div className="flex gap-2">
            {PrivacyPolicies.map((option: PrivacyPolicy) => (
              <button
                key={option.value}
                onClick={() => handleVisibilityChange(option.value)}
                className={`flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors ${
                  metadata.visibility === option.value
                    ? 'bg-blue-500 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                <option.icon className="w-4 h-4" />
                {option.label}
              </button>
            ))}
          </div>
        </div>

        {/* Email Access - Only show for private visibility */}
        <div className={`space-y-2 transition-all duration-300 ease-in-out overflow-hidden ${
          metadata.visibility === 'private'
            ? 'max-h-96 opacity-100'
            : 'max-h-0 opacity-0'
        }`}>
          <div className="flex items-center gap-2">
            <Mail className="w-4 h-4 text-gray-500" />
            <span className="text-sm font-medium text-gray-700">Email Access</span>
          </div>

          <div className="flex gap-2">
            <input
              type="email"
              value={emailInput}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setEmailInput(e.target.value)}
              onKeyPress={(e: KeyboardEvent<HTMLInputElement>) => e.key === 'Enter' && addEmail()}
              placeholder="Add email addresses"
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              onClick={addEmail}
              className="px-3 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 transition-colors"
            >
              <Plus className="w-4 h-4" />
            </button>
          </div>

          {metadata.emails.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {metadata.emails.map((email, index) => (
                <span
                  key={index}
                  onClick={() => removeEmail(email)}
                  className="inline-flex items-center px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full cursor-pointer hover:bg-green-200 transition-colors max-w-48 truncate"
                >
                  <span className="truncate">{email}</span>
                  <X className="w-3 h-3 ml-1 flex-shrink-0" />
                </span>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

// Main UploadCard component
const UploadCard: React.FC = () => {
  const [filesWithMetadata, setFilesWithMetadata] = useState<FileWithMetadata[]>([]);
  const [dragActive, setDragActive] = useState<boolean>(false);
  const [isUploading, setIsUploading] = useState<boolean>(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const navigate = useNavigate();
  const { req } = useAuth();

  const handleDrag = (e: DragEvent<HTMLDivElement>): void => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = (e: DragEvent<HTMLDivElement>): void => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    const droppedFiles = Array.from(e.dataTransfer.files);
    const newFilesWithMetadata = droppedFiles.map(file => ({
      file,
      metadata: {
        tags: [],
        visibility: 'anyone' as UploadVisibilityT,
        emails: [],
        description: ''
      }
    }));

    setFilesWithMetadata(prev => [...prev, ...newFilesWithMetadata]);
  };

  const handleFileSelect = (e: ChangeEvent<HTMLInputElement>): void => {
    if (e.target.files) {
      const selectedFiles = Array.from(e.target.files);
      const newFilesWithMetadata = selectedFiles.map(file => ({
        file,
        metadata: {
          tags: [],
          visibility: 'anyone' as UploadVisibilityT,
          emails: [],
          description: ''
        }
      }));

      setFilesWithMetadata(prev => [...prev, ...newFilesWithMetadata]);
    }
  };

  const removeFile = (fileToRemove: File): void => {
    setFilesWithMetadata(prev => prev.filter((item: FileWithMetadata) => item.file !== fileToRemove));
  };

  const clearAllFiles = (): void => {
    setFilesWithMetadata([]);
  };

  const updateFileMetadata = (file: File, metadata: FileMetadata): void => {
    setFilesWithMetadata(prev =>
      prev.map(item =>
        item.file === file
          ? { ...item, metadata }
          : item
      )
    );
  };

  const submitFiles = async (): Promise<void> => {
    if (filesWithMetadata.length === 0) {
      alert('No files to upload');
      return;
    }

    setIsUploading(true);

    try {
      // Create FormData
      const formData = new FormData();

      // Prepare metadata according to API format
      const metadata = filesWithMetadata.map(item => ({
        tags: item.metadata.tags,
        access: {
          level: item.metadata.visibility,
          allowed_user_emails: item.metadata.emails
        },
        description: item.metadata.description
      }));

      // Add metadata as first part
      formData.append('metadata', JSON.stringify({ metadata }));

      // Add files in the same order as metadata
      filesWithMetadata.forEach(item => {
        formData.append('file', item.file, item.file.name);
      });

      // Submit to API
      const response = await req(`/uploads`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();

      // Handle response
      console.log('Upload result:', result);

      // Check if all uploads were successful
      const failures = result.results?.filter((r: any) => !r.successful) || [];

      if (failures.length === 0) {
        alert(`Successfully uploaded ${filesWithMetadata.length} files!`);
        // Clear files after successful upload
        setFilesWithMetadata([]);
        // Optionally navigate away
        // navigate('/');
      } else {
        alert(`Upload completed with ${failures.length} failures. Check console for details.`);
        console.error('Failed uploads:', failures);
      }

    } catch (error) {
      console.error('Upload error:', error);
      alert('Upload failed. Please try again.');
    } finally {
      setIsUploading(false);
    }
  };

  const handleSubmit = (): void => {
    submitFiles();
  };

  const handleCancel = (): void => {
    navigate('/');
  };

  return (
    <div className="max-w-4xl mx-auto p-6 bg-white">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Upload Files</h1>
        <p className="text-gray-600">Select and configure files for upload</p>
      </div>

      {/* Dropzone */}
      <div
        className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
          dragActive
            ? 'border-blue-500 bg-blue-50'
            : 'border-gray-300 hover:border-gray-400'
        }`}
        onDragEnter={handleDrag}
        onDragLeave={handleDrag}
        onDragOver={handleDrag}
        onDrop={handleDrop}
      >
        <Upload className="w-12 h-12 text-gray-400 mx-auto mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          Drop files here or browse
        </h3>
        <p className="text-gray-500 mb-4">Support for multiple file types</p>

        <input
          ref={fileInputRef}
          type="file"
          multiple
          onChange={handleFileSelect}
          className="hidden"
        />

        <button
          type="button"
          onClick={() => fileInputRef.current?.click()}
          className="bg-blue-500 text-white px-6 py-2 rounded-md hover:bg-blue-600 transition-colors"
          disabled={isUploading}
        >
          Select Files
        </button>
      </div>

      {/* Files Summary */}
      {filesWithMetadata.length > 0 && (
        <div className="mt-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">
              Files to upload ({filesWithMetadata.length})
            </h2>
            <button
              type="button"
              onClick={clearAllFiles}
              className="text-red-500 hover:text-red-700 text-sm font-medium transition-colors"
              disabled={isUploading}
            >
              Clear all files
            </button>
          </div>

          {/* Files List */}
          <div className="space-y-4">
            {filesWithMetadata.map((item, index) => (
              <FilePreview
                key={`${item.file.name}-${index}`}
                file={item.file}
                metadata={item.metadata}
                onRemove={removeFile}
                onUpdateMetadata={updateFileMetadata}
              />
            ))}
          </div>
        </div>
      )}

      <div className="flex items-center justify-between">
        {/* Cancel Button */}
        <div className="mt-6 text-center">
          <button
            type="button"
            onClick={handleCancel}
            className="bg-red-400 text-white px-8 py-3 rounded-md text-lg font-medium hover:bg-red-500 transition-colors disabled:opacity-50"
            disabled={isUploading}
          >
            Cancel
          </button>
        </div>

        {/* Submit Button */}
        {filesWithMetadata.length > 0 && (
          <div className="mt-6 text-center">
            <button
              type="button"
              onClick={handleSubmit}
              className="bg-green-400 text-white px-8 py-3 rounded-md text-lg font-medium hover:bg-green-500 transition-colors disabled:opacity-50"
              disabled={isUploading}
            >
              {isUploading ? 'Uploading...' : 'Upload Files'}
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default UploadCard;