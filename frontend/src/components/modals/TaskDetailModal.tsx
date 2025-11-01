import React, { useState } from 'react';
import { X, Calendar, Tag, MessageSquare, Send } from 'lucide-react';
import { useTheme } from '../../contexts/ThemeContext';
import { Task, TaskComment } from '../../types';

interface TaskDetailModalProps {
  task: Task;
  onClose: () => void;
}

const TaskDetailModal: React.FC<TaskDetailModalProps> = ({ task, onClose }) => {
  const { theme } = useTheme();
  const [comments, setComments] = useState<TaskComment[]>([
    { id: 1, author: 'KIM', content: 'Looking good!', timestamp: '2h ago' },
    { id: 2, author: 'LEE', content: 'Need more details', timestamp: '1h ago' },
  ]);
  const [newComment, setNewComment] = useState('');

  const handleAddComment = () => {
    if (newComment.trim()) {
      setComments([
        ...comments,
        {
          id: comments.length + 1,
          author: 'PLAYER1',
          content: newComment,
          timestamp: 'Just now',
        },
      ]);
      setNewComment('');
    }
  };

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50 overflow-y-auto"
      onClick={onClose}
    >
      <div className="relative w-full max-w-2xl my-8" onClick={(e) => e.stopPropagation()}>
        <div
          className={`relative ${theme.colors.card} ${theme.effects.borderWidth} ${theme.colors.border} p-4 sm:p-6 max-h-[90vh] overflow-y-auto ${theme.effects.borderRadius} shadow-xl`}
        >
          <div
            className={`flex items-start justify-between mb-4 pb-4 ${theme.effects.borderWidth} ${theme.colors.border} border-t-0 border-l-0 border-r-0`}
          >
            <div className="flex-1 pr-4">
              <h2 className={`${theme.font.size.base} font-bold mb-2 break-words`}>{task.title}</h2>
              <div className="flex items-center gap-2 mt-3">
                <div
                  className={`w-8 h-8 ${theme.colors.primary} ${theme.effects.cardBorderWidth} ${theme.colors.border} flex items-center justify-center text-white ${theme.font.size.xs} font-bold ${theme.effects.borderRadius}`}
                >
                  {task.assignee[0]}
                </div>
                <span className={theme.font.size.xs}>{task.assignee}</span>
              </div>
            </div>
            <button
              onClick={onClose}
              className={`bg-red-500 ${theme.effects.cardBorderWidth} ${theme.colors.border} p-2 hover:bg-red-600 flex-shrink-0 ${theme.effects.borderRadius} transition`}
            >
              <X className="w-4 h-4 text-white" />
            </button>
          </div>

          <div className="space-y-4 mb-6">
            <div>
              <label
                className={`flex items-center gap-2 ${theme.font.size.xs} mb-2 ${theme.colors.textSecondary}`}
              >
                <Calendar className="w-4 h-4" />
                목표일 :
              </label>
              <div
                className={`px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} bg-gray-50 ${theme.font.size.xs} ${theme.effects.borderRadius}`}
              >
                2025-10-25
              </div>
            </div>

            <div>
              <label
                className={`flex items-center gap-2 ${theme.font.size.xs} mb-2 ${theme.colors.textSecondary}`}
              >
                <Tag className="w-4 h-4" />
                우선 순위 :
              </label>
              <span
                className={`inline-block px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} bg-red-500 text-white ${theme.font.size.xs} ${theme.effects.borderRadius}`}
              >
                HIGH
              </span>
            </div>

            <div>
              <label className={`${theme.font.size.xs} mb-2 ${theme.colors.textSecondary} block`}>
                설명:
              </label>
              <textarea
                className={`w-full px-3 py-2 ${theme.effects.cardBorderWidth} ${theme.colors.border} bg-gray-50 ${theme.font.size.xs} min-h-24 ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
                defaultValue="This is a detailed description of the task."
              />
            </div>
          </div>

          <div
            className={`${theme.effects.borderWidth} ${theme.colors.border} border-b-0 border-l-0 border-r-0 pt-4`}
          >
            <div className="flex items-center gap-2 mb-4">
              <MessageSquare className="w-4 h-4" />
              <h3 className={`${theme.font.size.xs} font-bold`}>댓글 ({comments.length})</h3>
            </div>

            <div className="space-y-3 mb-4 max-h-60 overflow-y-auto">
              {comments.map((comment) => (
                <div
                  key={comment.id}
                  className={`${theme.colors.card} ${theme.effects.cardBorderWidth} ${theme.colors.border} p-3 ${theme.effects.borderRadius}`}
                >
                  <div className="flex items-start gap-2">
                    <div
                      className={`w-6 h-6 ${theme.colors.primary} ${theme.effects.cardBorderWidth} ${theme.colors.border} flex items-center justify-center text-white text-xs font-bold flex-shrink-0 ${theme.effects.borderRadius}`}
                    >
                      {comment.author[0]}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-xs font-bold">{comment.author}</span>
                        <span className={`text-xs ${theme.colors.textSecondary}`}>
                          {comment.timestamp}
                        </span>
                      </div>
                      <p className={`${theme.font.size.xs} break-words`}>{comment.content}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            <div
              className={`${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.card} p-2 flex gap-2 ${theme.effects.borderRadius}`}
            >
              <input
                type="text"
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleAddComment()}
                placeholder="ADD COMMENT..."
                className={`flex-1 px-2 py-1 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.font.size.xs} ${theme.effects.borderRadius} focus:outline-none focus:ring-2 focus:ring-blue-500`}
              />
              <button
                onClick={handleAddComment}
                className={`${theme.colors.primary} text-white px-3 py-1 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.primaryHover} transition flex items-center gap-1 ${theme.effects.borderRadius}`}
              >
                <Send className="w-3 h-3" />
              </button>
            </div>
          </div>

          <div
            className={`flex gap-2 mt-6 pt-4 ${theme.effects.borderWidth} ${theme.colors.border} border-b-0 border-l-0 border-r-0`}
          >
            <button
              className={`flex-1 ${theme.colors.primary} text-white py-3 ${theme.effects.cardBorderWidth} ${theme.colors.border} ${theme.colors.primaryHover} transition ${theme.font.size.xs} ${theme.effects.borderRadius}`}
            >
              SAVE
            </button>
            <button
              className={`bg-red-500 text-white px-4 py-3 ${theme.effects.cardBorderWidth} ${theme.colors.border} hover:bg-red-600 transition ${theme.font.size.xs} ${theme.effects.borderRadius}`}
            >
              DELETE
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TaskDetailModal;
