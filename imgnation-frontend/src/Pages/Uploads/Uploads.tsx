import React, { useState, useEffect } from 'react';
import { Download, Heart, ThumbsDown, Eye, File, Image, FileText, Video, Music } from 'lucide-react';
import {useAuth} from "../../Providers/Api.tsx";
import {ArrowLeft} from "npm:lucide-react@0.514.0";
import { NavLink, useNavigate } from 'react-router';

interface Variant {
  name: string;
  content_type: string;
  size: number;
  metadata?: {
    height?: number;
    width?: number;
  };
  is_chunked: boolean;
  status: string;
}

interface Upload {
  user_id: string;
  filename: string;
  access: {
    level: number;
  };
  uploaded_at: string;
  tags: string[];
  description: string;
}

interface UploadData {
  id: string;
  size: number;
  variants: Variant[];
  uploads: Upload[];
  all_tags: string[];
  reactions: { [key: string]: number };
  your_reactions: string[];
  description: string;
  created_at: string;
}

interface ReactionResp {
  is_set: boolean;
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const getFileIcon = (contentType: string) => {
  if (contentType.startsWith('image/')) return <Image className="w-6 h-6" />;
  if (contentType.startsWith('video/')) return <Video className="w-6 h-6" />;
  if (contentType.startsWith('audio/')) return <Music className="w-6 h-6" />;
  if (contentType.includes('text') || contentType.includes('document')) return <FileText className="w-6 h-6" />;
  return <File className="w-6 h-6" />;
};

interface UploadViewPageProps {
  uploadId?: string;
}

const UploadViewPage: React.FC<UploadViewPageProps> = ({ uploadId: propUploadId }) => {
  const getUploadIdFromUrl = () => {
    if (globalThis !== undefined) {
      const urlParams = new URLSearchParams(globalThis.location.search);
      return urlParams.get('id');
    }
    return null;
  };

  const uploadId = propUploadId || getUploadIdFromUrl();

  const [uploadData, setUploadData] = useState<UploadData | null>(null);
  const [uploadLiked, setUploadLiked] = useState<boolean>(false);
  const [uploadDisliked, setUploadDisliked] = useState<boolean>(false);
  const [likesCount, setLikesCount] = useState<number | null>(null);
  const [dislikesCount, setDislikesCount] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [userInteractions, setUserInteractions] = useState({
    liked: false,
    disliked: false,
    viewed: false
  });

  const { req } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (!uploadId || uploadId === 'undefined') {
      setError('Upload ID is required');
      setLoading(false);
      return;
    }

    fetchUploadData().then(() => {});
  }, [uploadId]);

  const fetchUploadData = async () => {
    try {
      setLoading(true);
      console.log('Fetching data for uploadId:', uploadId); // Debug log
      const response = await req(`/uploads/data/${uploadId}`, {});

      if (!response.ok) {
        throw new Error(`Failed to fetch upload data: ${response.statusText}`);
      }

      const data: UploadData = await response.json();
      console.log('Received data:', data); // Debug log
      setLikesCount(data.reactions["like"] || 0)
      setDislikesCount(data.reactions["dislike"] || 0)
      setUploadLiked(data.your_reactions.includes("like"))
      setUploadDisliked(data.your_reactions.includes("dislike"))
      setUploadData(data);

      loadPreview(data);
    } catch (err) {
      console.error('Fetch error:', err); // Debug log
      setError(err instanceof Error ? err.message : 'Failed to load upload data');
    } finally {
      setLoading(false);
    }
  };

  const loadPreview = (data: UploadData) => {
    // Find the best preview variant in order of preference
    const variants = data.variants || [];
    const previewVariant =
      variants.find(v => v.name === 'compressed' && v.content_type.startsWith('image/')) ||
      variants.find(v => v.name === '' && v.content_type.startsWith('image/')) ||
      variants.find(v => v.name === 'thumbnail' && v.content_type.startsWith('image/'))

    if (previewVariant) {

      const variantParam = previewVariant.name ? `?variant=${previewVariant.name}` : '';
      req(`/uploads/${data.id}${variantParam}`, {})
        .then(res => res.blob())
        .then(blob => {
          setPreviewUrl(URL.createObjectURL(blob));
        });
    }
  };

  const handleDownload = async () => {
    if (!uploadData) return;
    try {
      const response = await req(`/uploads/${uploadData.id}`, {});

      if (!response.ok) {
        throw new Error('Failed to download file');
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `download-${uploadData.id}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err) {
      console.error('Download failed:', err);
      alert('Failed to download file');
    }
  };

  const handleLike = async () => {
    if (!uploadData) return;
    try {
      const resp: ReactionResp = await (await req(`/uploads/react/${uploadData.id}`, {
        method: 'POST',
        body: JSON.stringify({reaction: 'like'}),
      })).json()
      setLikesCount((likesCount || 0) + (resp.is_set ? 1 : -1))
      setUploadLiked(resp.is_set)
    } catch (err) {
      console.error('Failed to toggle like:', err);
    }
  };

  const handleDislike = async () => {
    if (!uploadData) return;
    try {
      const resp: ReactionResp = await (await req(`/uploads/react/${uploadData.id}`, {
        method: 'POST',
        body: JSON.stringify({reaction: 'dislike'}),
      })).json()
      setUploadDisliked(resp.is_set)
      setDislikesCount((dislikesCount || 0) + (resp.is_set ? 1 : -1))
    } catch (err) {
      console.error('Failed to toggle dislike:', err);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading upload...</p>
        </div>
      </div>
    );
  }

  if (error || !uploadData) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="text-center">
          <div className="text-red-500 text-xl mb-4">⚠️</div>
          <h2 className="text-xl font-semibold text-gray-800 mb-2">Error</h2>
          <p className="text-gray-600">{error || 'Upload not found'}</p>
        </div>
      </div>
    );
  }

  const originalVariant = uploadData.variants?.find(v => v.name === '');

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-4xl mx-auto px-4">
        <div className="bg-white rounded-lg shadow-lg overflow-hidden">
          {/* Header Section */}
          <div className="p-6 flex items-center justify-between">
            <NavLink
              to="/"
              className="group flex items-center space-x-2 text-gray-600 hover:text-indigo-600 transition-all duration-200"
            >
              <div className="flex items-center justify-center w-9 h-9 rounded-full bg-white shadow-sm group-hover:shadow-md group-hover:bg-indigo-50 transition-all duration-200">
                <ArrowLeft className="h-4 w-4 group-hover:text-indigo-600" />
              </div>
              <span className="text-sm font-medium group-hover:text-indigo-600">
                                Home
                            </span>
            </NavLink>
            <h1 className="text-2xl font-bold text-gray-800 mb-2">
              {uploadData.uploads[0].filename}
            </h1>
            {/* Spacer for balance */}
            <div className="w-16"></div>
          </div>

          {/* Preview Section */}
          <div className="bg-gray-100 p-2 text-center">
            {previewUrl ? (
              <img
                src={previewUrl}
                alt="Upload preview"
                className="max-w-full mx-auto rounded-lg shadow-md"
              />
            ) : (
              <div className="flex flex-col items-center justify-center py-16">
                {getFileIcon(originalVariant?.content_type || 'application/octet-stream')}
                <p className="text-gray-500 mt-4">No preview available</p>
              </div>
            )}
          </div>

          {/* File Information */}
          <div className="p-6">
            <div className="mb-6">
              {uploadData.description && (
                <p className="text-gray-600 mb-4">{uploadData.description}</p>
              )}
            </div>

            {/* File Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="font-medium text-gray-700">File Type:</span>
                  <span className="text-gray-600">{originalVariant?.content_type || 'Unknown'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="font-medium text-gray-700">Size:</span>
                  <span className="text-gray-600">{formatFileSize(uploadData.size || 0)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="font-medium text-gray-700">Uploaded:</span>
                  <span className="text-gray-600">
                    {new Date(uploadData.created_at).toLocaleDateString()}
                  </span>
                </div>
              </div>

              <div className="space-y-3">
                {originalVariant?.metadata && (
                  <div className="flex justify-between">
                    <span className="font-medium text-gray-700">Dimensions:</span>
                    <span className="text-gray-600">
                      {originalVariant.metadata.width} × {originalVariant.metadata.height}
                    </span>
                  </div>
                )}
              </div>
            </div>

            {/* Actions */}
            <div className="flex flex-wrap gap-4 items-center justify-between border-t pt-6">
              <div className="flex gap-4">
                <button
                  onClick={handleLike}
                  className={`flex items-center gap-2 px-4 py-2 rounded-lg transition-colors ${
                    uploadLiked 
                      ? 'bg-red-100 text-red-700 hover:bg-red-200'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  <Heart className={`w-4 h-4 ${userInteractions.liked ? 'fill-current' : ''}`} />
                  <span>{likesCount}</span>
                </button>

                <button
                  onClick={handleDislike}
                  className={`flex items-center gap-2 px-4 py-2 rounded-lg transition-colors ${
                    uploadDisliked
                      ? 'bg-blue-200 text-blue-800 hover:bg-blue-300'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  <ThumbsDown className={`w-4 h-4 ${userInteractions.disliked ? 'fill-current' : ''}`} />
                  <span>{dislikesCount}</span>
                </button>

                <div className="flex items-center gap-2 px-4 py-2 rounded-lg text-gray-700">
                  <Eye className="w-4 h-4" />
                  <span>{uploadData.reactions["view"] || 0}</span>
                </div>
              </div>

              <button
                onClick={handleDownload}
                className="flex items-center gap-2 px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
              >
                <Download className="w-4 h-4" />
                Download Original
              </button>
            </div>

            {/* Tags */}
            {uploadData.all_tags && uploadData.all_tags.length > 0 && (
              <div className="mt-6 pt-4 border-t" >
                <h3 className="font-medium text-gray-700 mb-2">Tags</h3>
                <div className="flex flex-wrap gap-2">
                  {uploadData.all_tags.map((tag, index) => (
                    <span
                      key={index}
                      className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm"
                      onClick={() => navigate('/search?tags=' + tag)}
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default UploadViewPage;